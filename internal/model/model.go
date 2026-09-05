// Package model содержит именованные типы запросов и ответов API,
// чтобы и сервер, и клиенты ссылались на одно определение.
package model

// ShortenRequest — запрос на создание короткой ссылки.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse — ответ с созданной короткой ссылкой.
type ShortenResponse struct {
	SlugURL string `json:"result"`
}
