package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/liuzl/gocc"
	"github.com/mmcdole/gofeed"
	"log"
	"net/http"
	"strings"
	"time"
)

// NewsResponse 定义新闻响应的JSON结构
type NewsResponse struct {
	Title      string   `json:"title"`
	Time       string   `json:"time"`
	Content    string   `json:"content"`
	ImageURLs  []string `json:"image_urls"`
	Paragraphs []string `json:"paragraphs"`
	HTML       string   `json:"html"` // 新增HTML字段
}

// TopNewsResponse 定义头条新闻响应的JSON结构
type TopNewsResponse struct {
	Title      string   `json:"title"`
	Time       string   `json:"time"`
	Content    string   `json:"content"`
	ImageURLs  []string `json:"image_urls"`
	Paragraphs []string `json:"paragraphs"`
	URL        string   `json:"url"`
	HTML       string   `json:"html"` // 新增HTML字段
}

func main() {
	router := gin.Default()

	// 获取头条新闻列表
	router.GET("/", func(c *gin.Context) {
		news, err := getTopNews()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 为每个新闻获取图片
		var newsWithImages []gin.H
		for _, item := range news {
			url := item["url"].(string)
			_, imageURLs, err := getNewsContent(url)
			if err != nil {
				log.Printf("获取新闻图片失败: %v", err)
				// 如果获取图片失败，使用空字符串
				item["image_url"] = ""
			} else if len(imageURLs) > 0 {
				// 只取第一张图片
				item["image_url"] = imageURLs[0]
			} else {
				item["image_url"] = ""
			}
			newsWithImages = append(newsWithImages, item)
		}

		c.JSON(http.StatusOK, newsWithImages)
	})

	// 获取头条新闻列表（包含内容）
	router.GET("/api/news/top", func(c *gin.Context) {
		news, err := getTopNews()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		var fullNews []TopNewsResponse
		for _, item := range news {
			url := item["url"].(string)
			title := item["title"].(string)
			timeStr := item["time"].(string)

			content, imageURLs, err := getNewsContent(url)
			if err != nil {
				log.Printf("获取新闻内容失败: %v", err)
				continue
			}

			// 将内容分段
			paragraphs := strings.Split(content, "\n")
			// 过滤空段落
			var filteredParagraphs []string
			for _, p := range paragraphs {
				if strings.TrimSpace(p) != "" {
					filteredParagraphs = append(filteredParagraphs, strings.TrimSpace(p))
				}
			}

			fullNews = append(fullNews, TopNewsResponse{
				Title:      title,
				Time:       timeStr,
				Content:    content,
				ImageURLs:  imageURLs,
				Paragraphs: filteredParagraphs,
				URL:        url,
				HTML:       generateHTML(title, timeStr, content, imageURLs),
			})
		}

		c.JSON(http.StatusOK, fullNews)
	})

	// 获取指定URL的新闻内容
	router.GET("/api/news/content", func(c *gin.Context) {
		url := c.Query("url")
		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL参数不能为空",
			})
			return
		}

		content, imageURLs, err := getNewsContent(url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 将内容按段落分割
		paragraphs := strings.Split(content, "\n")
		// 过滤空段落
		var filteredParagraphs []string
		for _, p := range paragraphs {
			if strings.TrimSpace(p) != "" {
				filteredParagraphs = append(filteredParagraphs, p)
			}
		}

		// 获取当前时间
		t := time.Now()

		// 生成HTML内容
		htmlContent := generateHTML(filteredParagraphs[0], t.Format("2006年01月02日 15:04:05"), content, imageURLs)

		// 根据Accept头决定返回格式
		accept := c.GetHeader("Accept")
		if strings.Contains(accept, "text/html") {
			// 如果客户端请求HTML，直接返回HTML内容
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, htmlContent)
		} else if strings.Contains(accept, "application/json") {
			// 如果客户端请求JSON，返回JSON格式
			response := NewsResponse{
				Content:    content,
				ImageURLs:  imageURLs,
				Paragraphs: filteredParagraphs,
				Title:      filteredParagraphs[0],
				Time:       t.Format("2006年01月02日 15:04:05"),
				HTML:       htmlContent,
			}
			c.JSON(http.StatusOK, response)
		} else {
			// 默认返回未转义的HTML
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, htmlContent)
		}
	})

	// 获取新闻详情
	router.GET("/news", func(c *gin.Context) {
		url := c.Query("url")
		title := c.Query("title")
		timeStr := c.Query("time")

		if url == "" || title == "" || timeStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "缺少必要参数",
			})
			return
		}

		// 解析时间
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "时间格式错误",
			})
			return
		}

		content, imageURLs, err := getNewsContent(url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 将内容分段
		paragraphs := strings.Split(content, "\n")
		// 过滤空段落
		var filteredParagraphs []string
		for _, p := range paragraphs {
			if strings.TrimSpace(p) != "" {
				filteredParagraphs = append(filteredParagraphs, strings.TrimSpace(p))
			}
		}

		response := NewsResponse{
			Title:      title,
			Time:       t.Format("2006年01月02日 15:04:05"),
			Content:    content,
			ImageURLs:  imageURLs,
			Paragraphs: filteredParagraphs,
			HTML:       generateHTML(title, t.Format("2006年01月02日 15:04:05"), content, imageURLs),
		}

		c.JSON(http.StatusOK, response)
	})

	router.Run(":8001")
}

func getTopNews() ([]gin.H, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://feeds.bbci.co.uk/zhongwen/trad/rss.xml")
	if err != nil {
		return nil, err
	}

	// 创建繁体到简体的转换器
	cc, err := gocc.New("t2s")
	if err != nil {
		return nil, err
	}

	var news []gin.H
	// 只返回前3条新闻
	for i := 0; i < 3 && i < len(feed.Items); i++ {
		item := feed.Items[i]
		// 转换标题为简体中文
		simplifiedTitle, err := cc.Convert(item.Title)
		if err != nil {
			return nil, err
		}

		news = append(news, gin.H{
			"title": simplifiedTitle,
			"url":   item.Link,
			"time":  item.Published,
		})
	}

	return news, nil
}

func getNewsContent(url string) (string, []string, error) {
	// 创建繁体到简体的转换器
	t2s, err := gocc.New("t2s")
	if err != nil {
		log.Printf("创建繁简转换器失败: %v", err)
		return "", nil, fmt.Errorf("创建繁简转换器失败: %v", err)
	}

	// 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return "", nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return "", nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP请求失败，URL: %s, 状态码: %d", url, resp.StatusCode)
		return "", nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("解析HTML失败: %v", err)
		return "", nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	var content strings.Builder
	var imageURLs []string

	// BBC中文网文章标题选择器
	titleSelectors := []string{
		"h1[tabindex='-1']",     // 新版文章标题
		"h1.story-body__h1",     // 旧版文章标题
		"h1.bbc-title-lg",       // 另一种标题格式
		".article-headline__text", // 新闻标题
		".story-body h1",         // 通用标题
		"article h1",             // 通用标题2
		".article__header h1",    // 文章头部标题
		"#main-heading",          // 主标题
		".vxp-media__headline",   // 视频新闻标题
	}

	// 尝试获取标题
	title := ""
	for _, selector := range titleSelectors {
		title = strings.TrimSpace(doc.Find(selector).First().Text())
		if title != "" {
			content.WriteString(title + "\n\n")
			log.Printf("成功提取标题: %s", title)
			break
		}
	}

	// BBC中文网图片选择器
	imageSelectors := []string{
		".article__body-content img",                // 文章内容图片
		".story-body__inner img",                    // 旧版文章内容图片
		".article-body-container img",               // 通用文章内容图片
		".body-content-container img",               // 正文容器图片
		"[data-component='image-block'] img",        // 图片块组件
		".article__body img",                        // 文章主体图片
		".body-text-card img",                       // 文本卡片中的图片
		".article-body__image-container img",        // 文章图片容器
		"figure.image-block img",                    // 图片块
		"figure.media-with-caption img",             // 带标题的媒体
		".js-image-replace",                         // 延迟加载的图片

		"figure img.bbc-image",                      // BBC标准图片
		"figure img.js-image-replace",               // 旧版图片
		"figure img.sp-media-asset_img",             // 体育新闻图片
		".article__inline-image img",                // 内联图片
		".article-figure__image img",                // 文章图片
		".image-block img",                          // 图片块
		".vxp-media__player img",                    // 视频缩略图
		".image-and-copyright-container img",        // 带版权信息的图片

		".js-delayed-image-load",                    // 延迟加载的图片
		"picture source",                            // picture元素的source
		".responsive-image img",                     // 响应式图片
		"img[data-src]",                            // 带data-src属性的图片
		"img[data-delayed-src]",                    // 带data-delayed-src属性的图片

		"meta[property='og:image']",                // Open Graph图片
		"meta[name='twitter:image']",               // Twitter卡片图片

		"article img",                              // 文章中的所有图片
		".content img",                             // 内容区域的所有图片
		"main img",                                 // 主要内容区域的所有图片
	}

	// 用于去重的map
	seenURLs := make(map[string]bool)

	// 尝试获取所有图片URL
	for _, selector := range imageSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			// 检查多个可能的属性
			possibleAttrs := []string{
				"data-src",           // 延迟加载的原始图片
				"src",                // 标准src属性
				"data-delayed-src",   // 延迟加载
				"data-original",      // 原始图片
				"data-highres",       // 高分辨率图片
				"content",            // meta标签的content属性
				"srcset",             // 响应式图片集
			}

			for _, attr := range possibleAttrs {
				if val, exists := s.Attr(attr); exists && val != "" {
					var imgURL string
					// 处理srcset属性
					if attr == "srcset" {
						srcsetParts := strings.Split(val, ",")
						if len(srcsetParts) > 0 {
							// 获取第一个URL（通常是最大尺寸）
							firstSrcset := strings.Split(strings.TrimSpace(srcsetParts[0]), " ")[0]
							if firstSrcset != "" {
								imgURL = firstSrcset
							}
						}
					} else {
						// 处理相对URL
						if strings.HasPrefix(val, "//") {
							imgURL = "https:" + val
						} else if !strings.HasPrefix(val, "http") && !strings.HasPrefix(val, "/") {
							imgURL = "https://www.bbc.com/" + val
						} else if strings.HasPrefix(val, "/") {
							imgURL = "https://www.bbc.com" + val
						} else {
							imgURL = val
						}
					}

					// 如果找到有效的URL且未重复，则添加到列表中
					if imgURL != "" && !seenURLs[imgURL] {
						// 验证URL是否为图片URL
						if strings.Contains(imgURL, ".jpg") ||
							strings.Contains(imgURL, ".jpeg") ||
							strings.Contains(imgURL, ".png") ||
							strings.Contains(imgURL, ".gif") ||
							strings.Contains(imgURL, ".webp") {
							seenURLs[imgURL] = true
							imageURLs = append(imageURLs, imgURL)
							log.Printf("成功提取图片URL: %s，使用选择器: %s", imgURL, selector)
						}
					}
					break
				}
			}
		})
	}

	// BBC中文网文章内容选择器
	contentSelectors := []string{
		"article[role='main'] p",           // 新版文章内容
		".story-body__inner p",             // 旧版文章内容
		".article__body-content p",         // 另一种文章内容
		".article-body-container p",        // 通用文章内容
		"[data-component='text-block'] p",  // 文本块
		".bbc-19j92fr p",                  // 特殊格式
		".story-body__inner div.mapped-include p", // 嵌入内容
		".body-content-container p",        // 正文容器
		".article__body p",                 // 文章主体
		".vxp-media__summary p",           // 视频描述
	}

	// 尝试获取文章内容
	contentFound := false
	paragraphCount := 0
	for _, selector := range contentSelectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				content.WriteString(text + "\n")
				paragraphCount++
				contentFound = true
			}
		})
		if contentFound {
			log.Printf("使用选择器 '%s' 成功提取了 %d 段内容", selector, paragraphCount)
			break
		}
	}

	// 如果没有找到内容，尝试更通用的选择器
	if !contentFound {
		log.Printf("使用通用选择器尝试提取内容")
		doc.Find("article").Find("p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				content.WriteString(text + "\n")
				paragraphCount++
			}
		})
	}

	// 将内容转换为简体中文
	contentStr := content.String()
	contentStr, err = t2s.Convert(contentStr)
	if err != nil {
		log.Printf("繁体转简体失败: %v", err)
		return "", nil, fmt.Errorf("繁体转简体失败: %v", err)
	}

	// 如果内容为空，返回错误
	if contentStr == "" {
		log.Printf("无法提取文章内容，URL: %s", url)
		return "", nil, fmt.Errorf("无法提取文章内容")
	}

	log.Printf("成功提取文章内容，共 %d 段，找到 %d 张图片", paragraphCount, len(imageURLs))
	return contentStr, imageURLs, nil
}

// 生成HTML内容
func generateHTML(title, time string, content string, imageURLs []string) string {
	var html strings.Builder
	html.WriteString(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + title + `</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f8f8f8;
        }
        .article {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            padding: 30px;
            margin-bottom: 20px;
        }
        .title {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 10px;
            color: #1a1a1a;
            text-align: center;
        }
        .meta {
            font-size: 14px;
            color: #888;
            margin-bottom: 20px;
            text-align: center;
        }
        .content {
            font-size: 16px;
            line-height: 1.8;
            color: #333;
        }
        .content p {
            margin-bottom: 16px;
            text-align: justify;
        }
        .image-container {
            margin: 20px 0;
            text-align: center;
        }
        .image-container img {
            max-width: 100%;
            height: auto;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }
        .copyright {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            font-size: 12px;
            color: #999;
            text-align: center;
        }
        @media (max-width: 600px) {
            body {
                padding: 15px;
            }
            .article {
                padding: 20px;
            }
            .title {
                font-size: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="article">
        <h1 class="title">` + title + `</h1>
        <div class="meta">` + time + `</div>
        <div class="content">`)

	// 将内容按段落分割
	paragraphs := strings.Split(content, "\n")
	imageIndex := 0
	paragraphCount := 0

	// 添加第一张图片作为头图
	if len(imageURLs) > 0 {
		html.WriteString(`
            <div class="image-container">
                <img src="` + imageURLs[imageIndex] + `" alt="头图">
            </div>`)
		imageIndex++
	}

	// 处理每个段落
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		html.WriteString("<p>" + p + "</p>")
		paragraphCount++

		// 每隔5-7个段落插入一张图片（如果还有图片的话）
		if imageIndex < len(imageURLs) && paragraphCount >= 5 && paragraphCount <= 7 {
			html.WriteString(`
            <div class="image-container">
                <img src="` + imageURLs[imageIndex] + `" alt="配图">
            </div>`)
			imageIndex++
			paragraphCount = 0 // 重置段落计数
		}
	}

	// 添加版权信息
	html.WriteString(`
        </div>
        <div class="copyright">
            <p>本文内容来源于BBC中文网，如有侵权请联系必删</p>
            <p>Copyright ` + time[:4] + ` BBC. 保留所有权利。</p>
        </div>
    </div>
</body>
</html>`)

	return html.String()
}
