package parse

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"vacancies/config"
	"vacancies/grpc"

	"github.com/PuerkitoBio/goquery"
)

const startPage = "/vacancies?type=all"

func Do(config *config.Config, authorizationData grpc.AuthorizationData) error {
	client := &http.Client{}

	nickname, err := checkAuthorization(
		config,
		client,
		authorizationData,
	)
	if err != nil {
		return fmt.Errorf("checkAuthorization: %w", err)
	}
	log.Println("Successfully authorized with nickname:", nickname[1:])

	log.Println("Start parsing vacancies by keyword:", config.Keyword)
	// цикл прохода по страничкам с вакансиями(пагинация)
	vacanciesPage, vacanciesPageCount, vacanciesCount := startPage, 1, 0
	for {
		vacanciesPage = config.Domain + vacanciesPage
		vacanciesDocument, parseErr := parseHTMLDocument(
			config,
			client,
			vacanciesPage,
			authorizationData.Cookies,
		)
		if err != nil {
			return fmt.Errorf("parseHTMLDocument: %w", parseErr)
		}

		var vacanciesPageErr error
		// собираем ссылки на вакансии, выборка не учитывает быстрые вакансии хабра
		// чтобы не повторятся, потому что они носят рекламный характер и постоянно меняются
		vacanciesDocument.
			Find("a.vacancy-card__title-link").
			Each(func(i int, s *goquery.Selection) {
				vacancyPage, exists := s.Attr("href")
				if !exists {
					vacanciesPageErr = fmt.Errorf(
						"failed to find any vacancy link on vacancies page: %s",
						vacanciesPage,
					)
					return
				}

				vacancyPage = config.Domain + vacancyPage
				// получаем HTML документ вакансии
				vacancyDocument, vacancyErr := parseHTMLDocument(
					config,
					client,
					vacancyPage,
					authorizationData.Cookies,
				)
				if vacancyErr != nil {
					vacanciesPageErr = fmt.Errorf(
						"failed to parseHTMLDocument on page: %s err: %w",
						vacanciesPage,
						vacancyErr,
					)
					return
				}

				// проверяем страницу вакансии на наличие keyword-a в тексте
				if isKeywordInVacancy(vacancyDocument, strings.ToLower(config.Keyword)) {
					// TODO: в очередь
					fmt.Printf("%s -- %s\n", s.Text(), vacancyPage)
				}
				i++
				vacanciesCount++
			})

		if vacanciesPageErr != nil {
			return vacanciesPageErr
		}

		/// получаем ссылку следующей страницы пока не наткнемся на аттрибут disabled
		// что означает, что мы достигли последней страницы вакансий
		nextPage, pageExist := getElementAttribute(
			vacanciesDocument,
			"div a[rel='next']:not([disabled])",
			"href",
		)
		if !pageExist {
			log.Println("Next page didnt found, last page:", vacanciesPage)
			log.Println("Pages with vacancies visited:", vacanciesPageCount)
			log.Println("Parsed vacancies count:", vacanciesCount)
			break
		}

		vacanciesPage = nextPage
		vacanciesPageCount++
	}

	return nil
}

// isKeywordInVacancy - возвращает true если в вакансии найденно
// ключевое слово - keyword, иначе false
func isKeywordInVacancy(doc *goquery.Document, keyword string) bool {
	if strings.Contains(makeStringBlob(doc), keyword) {
		return true
	}
	return false
}

// makeStringBlob - склеиваем все полученные строки из вакансии
func makeStringBlob(doc *goquery.Document) string {
	var stringsBlob strings.Builder

	// заголовок
	stringsBlob.WriteString(doc.Find("h1.page-title__title").Text())

	// разные теги о компании и вакансии
	doc.Find("a.link-comp").Each(func(i int, s *goquery.Selection) {
		stringsBlob.WriteString(s.Text())
	})
	stringsBlob.WriteString(doc.Find(".vacancy-company__sub-title").Text())

	// местоположение и тип занятости
	stringsBlob.WriteString(doc.Find("span.inline-list").Text())

	// описание вакансии/компании/ожидания от кандидата/условия/требования и тд
	doc.Find(".style-ugc").Each(func(i int, s *goquery.Selection) {
		stringsBlob.WriteString(s.Text())
	})

	// возвращаем весь текст приведенный к нижнему регистру
	return strings.ToLower(stringsBlob.String())
}

// getElementAttribute - получаем значение аттрибута у элемента
// по переданному селектору и названию аттрибута
func getElementAttribute(
	doc *goquery.Document,
	selector, attribute string,
) (string, bool) {
	return doc.Find(selector).Attr(attribute)
}

// parseHTMLDocument - делает запрос по урлу и возвращает HTML документ страницы
func parseHTMLDocument(
	config *config.Config,
	c *http.Client,
	URL, cookies string,
) (*goquery.Document, error) {
	resp, err := requestURL(config, c, URL, cookies)
	if err != nil {
		return nil, fmt.Errorf("requestURL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"request failed on url: %s\n Status code: %d",
			URL,
			resp.StatusCode,
		)
	}

	var body io.Reader = resp.Body
	// если получаем gzip, то распаковываем в обычный html
	if !resp.Uncompressed {
		body, err = decompressReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressResponse: %w", err)
		}
	}

	// парсим HTML document
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("goquery.NewDocumentFromReader: %w", err)
	}

	return doc, nil
}

// requestURL - делает запрос по переданному урлу и возвращает ответ запроса
func requestURL(
	config *config.Config,
	client *http.Client,
	URL, cookies string,
) (*http.Response, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	err = SetHeaders(req, config.Domain, cookies)
	if err != nil {
		return nil, fmt.Errorf("SetHeaders: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %w", err)
	}

	return resp, nil
}

func SetHeaders(r *http.Request, domain string, cookies string) error {
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	r.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	r.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Host", domain)
	r.Header.Set("Referer", domain+startPage)
	r.Header.Set("Sec-Fetch-Dest", "document")
	r.Header.Set("Sec-Fetch-Mode", "navigate")
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	r.Header.Set("Sec-Fetch-User", "?1")
	r.Header.Set("Upgrade-Insecure-Requests", "1")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 8.0.0; SM-G955U Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Mobile Safari/537.36")
	r.Header.Set("Sec-Ch-Ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	r.Header.Set("Sec-Ch-Ua-Mobile", "?1")
	r.Header.Set("Sec-Ch-Ua-Platform", `"Android"`)
	r.Header.Set("Cookie", cookies)

	return nil
}

// decompressReader функция для распаковки содержимого ответа из gzip
func decompressReader(reader io.Reader) (io.Reader, error) {
	// Создаем новый Reader, который распакует gzip-сжатые данные из reader
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader: %s", err)
	}
	return gz, nil
}

// checkAuthorization - находим ссылку на страницу авторизованного пользователя
// формата /никнеймпользователя таким образом проверяем авторизовались мы или нет
func checkAuthorization(
	config *config.Config,
	client *http.Client,
	authorizationData grpc.AuthorizationData,
) (string, error) {
	// получаем HTML документ
	doc, err := parseHTMLDocument(
		config,
		client,
		config.Domain+startPage,
		authorizationData.Cookies,
	)
	if err != nil {
		return "", fmt.Errorf("parseHTMLDocument: %w", err)
	}

	userPage, exist := getElementAttribute(doc, "a.menu-head", "href")
	if !exist {
		return "", fmt.Errorf("href attribute for signed user not found")
	}

	return userPage, nil
}
