package hikvision

import "encoding/xml"

// PlaybackSearchRequest - структура для поиска записей
type PlaybackSearchRequest struct {
	XMLName              xml.Name `xml:"CMSearchDescription"`
	SearchID             string   `xml:"searchID"`
	SearchResultPosition int      `xml:"searchResultPosition"`
	MaxResults           int      `xml:"maxResults"`
	SearchMode           string   `xml:"searchMode"`
	StartTime            string   `xml:"timeSpanList>timeSpan>startTime"`
	EndTime              string   `xml:"timeSpanList>timeSpan>endTime"`
	Channels             string   `xml:"channelList>channelId"`
}

// Recording - структура для хранения информации о записи
type Recording struct {
	StartTime string `xml:"timeSpan>startTime"`
	EndTime   string `xml:"timeSpan>endTime"`
	Channel   string `xml:"trackID"`
}

// SearchResponse - структура для ответа на поиск записей
type SearchResponse struct {
	MatchList struct {
		Recordings []Recording `xml:"searchMatchItem"`
	} `xml:"matchList"`
}
