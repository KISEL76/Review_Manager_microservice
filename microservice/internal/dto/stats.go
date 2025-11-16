package dto

type ReviewerStat struct {
	Username     string `json:"username"`
	ReviewsCount int    `json:"reviews_count"`
}

/* /stats/reviewers */
type ReviewerStatsResponse struct {
	Stats []ReviewerStat `json:"stats"`
}
