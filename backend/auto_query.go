// 自动查询验证码（后台定时任务）
// 每30秒查询一次未获取验证码的卡密
func startAutoQuery() {
	// 检查是否开启自动查询
	if os.Getenv("AUTO_QUERY_ENABLED") != "true" {
		log.Println("自动查询已关闭，设置 AUTO_QUERY_ENABLED=true 开启")
		return
	}
	
	interval := 30
	if val := os.Getenv("AUTO_QUERY_INTERVAL"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			interval = i
		}
	}
	
	log.Printf("自动查询已启动，每%d秒执行一次", interval)
	
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for range ticker.C {
			autoQueryPendingCards()
		}
	}()
}

// 获取自动查询状态
func getAutoQueryStatus(c *gin.Context) {
	enabled := os.Getenv("AUTO_QUERY_ENABLED") == "true"
	interval := 30
	if val := os.Getenv("AUTO_QUERY_INTERVAL"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i > 0 {
			interval = i
		}
	}
	
	c.JSON(200, Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"enabled":  enabled,
			"interval": interval,
		},
	})
}

// 手动触发自动查询
func triggerAutoQuery(c *gin.Context) {
	go autoQueryPendingCards()
	c.JSON(200, Response{
		Code:    0,
		Message: "已触发自动查询",
	})
}

// 自动查询未获取验证码的卡密
func autoQueryPendingCards() {
	// 查询未获取验证码的卡密（限制每次查询10条，避免压力过大）
	rows, err := db.Query(
		"SELECT card_no, card_link FROM cards WHERE card_check = 0 AND card_link != '' LIMIT 10",
	)
	if err != nil {
		log.Printf("自动查询：查询卡密失败: %v", err)
		return
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var cardNo, cardLink string
		if err := rows.Scan(&cardNo, &cardLink); err != nil {
			continue
		}
		
		// 调用远程接口查询验证码
		go func(no, link string) {
			resp, err := http.Get(link)
			if err != nil {
				log.Printf("自动查询 %s: 远程接口错误", no)
				return
			}
			defer resp.Body.Close()
			
			var remoteResp RemoteResponse
			if err := json.NewDecoder(resp.Body).Decode(&remoteResp); err != nil {
				log.Printf("自动查询 %s: 解析响应失败", no)
				return
			}
			
			rawNote, _ := json.Marshal(remoteResp)
			note := string(rawNote)
			
			if remoteResp.Code == 1 && remoteResp.Data.Code != "" {
				code := extractVerificationCode(remoteResp.Data.Code)
				expired := convertTimeFormat(remoteResp.Data.ExpiredDate)
				_, err = db.Exec(
					"UPDATE cards SET card_code=?, card_expired_date=?, card_note=?, card_check=1 WHERE card_no=?",
					code, expired, note, no,
				)
				if err != nil {
					log.Printf("自动查询 %s: 更新数据库失败", no)
				} else {
					log.Printf("自动查询 %s: 成功获取验证码 %s", no, code)
				}
			} else {
				// 记录查询过，但没有获取到验证码
				_, err = db.Exec(
					"UPDATE cards SET card_note=?, card_check=1 WHERE card_no=?",
					note, no,
				)
				if err != nil {
					log.Printf("自动查询 %s: 标记已查失败", no)
				}
			}
		}(cardNo, cardLink)
		
		count++
	}
	
	if count > 0 {
		log.Printf("自动查询：本次处理 %d 条卡密", count)
	}
}