package e2e

import (
	"net/http"
	"testing"
)

/*
Полный E2E-сценарий:
1) /team/add — создаем команду backend с 4 участниками
2) /team/get — проверяем, что команда и участники читаются
3) /users/setIsActive — отключаем одного участника, чтобы его не назначало ревьювером
4) /pullRequest/create — создаем 2 PR от автора, проверяем автоназначение ревьюверов
5) /users/getReview — проверяем, что ревьювер видит свои PR
6) /pullRequest/reassign — переназначаем ревьювера
7) /pullRequest/merge — мержим PR и проверяем статус
8) /pullRequest/reassign (повторно) — убеждаемся, что для MERGED PR прилетает 409
9) /stats/reviewers — проверяем, что статистика по ревьюверам не пустая
*/
func TestE2E_FullFlow(t *testing.T) {
	teamName := "backend"

	// 1. /team/add — создаем команду
	var teamResp teamAddResponse

	doJSON(
		t,
		http.MethodPost,
		"/team/add",
		map[string]any{
			"team_name": teamName,
			"members": []map[string]any{
				{"user_id": "u1", "username": "Author", "is_active": true},
				{"user_id": "u2", "username": "Reviewer1", "is_active": true},
				{"user_id": "u3", "username": "Reviewer2", "is_active": true},
				{"user_id": "u4", "username": "Reviewer3", "is_active": true},
			},
		},
		http.StatusCreated,
		&teamResp,
	)

	if teamResp.Team.TeamName != teamName {
		t.Fatalf("ожидали team_name=%s, получили %s", teamName, teamResp.Team.TeamName)
	}
	if len(teamResp.Team.Members) != 4 {
		t.Fatalf("ожидали 4 участника в команде, получили %d", len(teamResp.Team.Members))
	}

	// 2. /team/get — проверяем, что команда читается корректно
	var teamGet teamGetResponse
	doJSON(
		t,
		http.MethodGet,
		"/team/get?team_name="+teamName,
		nil,
		http.StatusOK,
		&teamGet,
	)

	if teamGet.TeamName != teamName {
		t.Fatalf("team/get: ожидали team_name=%s, получили %s", teamName, teamGet.TeamName)
	}
	if len(teamGet.Members) != 4 {
		t.Fatalf("team/get: ожидали 4 участника, получили %d", len(teamGet.Members))
	}

	// 3. /users/setIsActive — отключаем u3 (не должен назначаться ревьювером)
	doJSON(
		t,
		http.MethodPost,
		"/users/setIsActive",
		map[string]any{
			"user_id":   "u3",
			"is_active": false,
		},
		http.StatusOK,
		nil,
	)

	// 4. Создаем два PR и проверяем автоназначение ревьюверов
	prIDs := []string{"pr-1", "pr-2"}
	var firstPR prResponse
	var firstReviewerID string

	for i, prID := range prIDs {
		var prResp prResponse
		doJSON(
			t,
			http.MethodPost,
			"/pullRequest/create",
			map[string]any{
				"pull_request_id":   prID,
				"pull_request_name": "E2E test " + prID,
				"author_id":         "u1",
			},
			http.StatusCreated,
			&prResp,
		)

		if prResp.PR.Status != "OPEN" {
			t.Fatalf("для %s ожидали статус OPEN, получили %s", prID, prResp.PR.Status)
		}

		revs := prResp.PR.AssignedReviewers
		if len(revs) == 0 || len(revs) > 2 {
			t.Fatalf("для %s ожидали 1 или 2 ревьювера, получили %d", prID, len(revs))
		}
		for _, rid := range revs {
			if rid == "u1" {
				t.Fatalf("автор не должен быть ревьювером, но для %s есть %s", prID, rid)
			}
			if rid == "u3" {
				t.Fatalf("пользователь u3 выключен, но его назначило ревьювером для %s", prID)
			}
		}

		if i == 0 {
			firstPR = prResp
			firstReviewerID = revs[0]
		}
	}

	if firstReviewerID == "" {
		t.Fatalf("ожидали хотя бы одного ревьювера у первого PR")
	}

	// 5. /users/getReview — проверяем, что первый ревьювер видит свой PR
	var reviewResp userGetReviewResponse
	doJSON(
		t,
		http.MethodGet,
		"/users/getReview?user_id="+firstReviewerID,
		nil,
		http.StatusOK,
		&reviewResp,
	)

	if reviewResp.UserID != firstReviewerID {
		t.Fatalf("users/getReview: ожидали user_id=%s, получили %s", firstReviewerID, reviewResp.UserID)
	}
	foundPR := false
	for _, pr := range reviewResp.PullRequests {
		if pr.PullRequestID == firstPR.PR.PullRequestID {
			foundPR = true
			break
		}
	}
	if !foundPR {
		t.Fatalf("users/getReview: не нашли PR %s в списке PR пользователя %s", firstPR.PR.PullRequestID, firstReviewerID)
	}

	// 6. /pullRequest/reassign — переназначаем первого ревьювера
	var reassignResp reassignResponse
	doJSON(
		t,
		http.MethodPost,
		"/pullRequest/reassign",
		map[string]any{
			"pull_request_id": firstPR.PR.PullRequestID,
			"old_user_id":     firstReviewerID,
		},
		http.StatusOK,
		&reassignResp,
	)

	if reassignResp.ReplacedBy == "" {
		t.Fatalf("ожидали, что replaced_by будет заполнен")
	}
	if reassignResp.ReplacedBy == firstReviewerID {
		t.Fatalf("ожидали нового ревьювера, отличного от старого, получили того же %s", firstReviewerID)
	}

	// 7. /pullRequest/merge — мержим первый PR
	var mergeResp prResponse
	doJSON(
		t,
		http.MethodPost,
		"/pullRequest/merge",
		map[string]any{
			"pull_request_id": firstPR.PR.PullRequestID,
		},
		http.StatusOK,
		&mergeResp,
	)

	if mergeResp.PR.Status != "MERGED" {
		t.Fatalf("после merge ожидали статус MERGED, получили %s", mergeResp.PR.Status)
	}

	// 8. Повторный /pullRequest/reassign для MERGED PR должен вернуть 409
	doJSON(
		t,
		http.MethodPost,
		"/pullRequest/reassign",
		map[string]any{
			"pull_request_id": firstPR.PR.PullRequestID,
			"old_user_id":     reassignResp.ReplacedBy,
		},
		http.StatusConflict,
		nil,
	)

	// 9. /stats/reviewers — статистика по ревьюверам
	var statsResp reviewerStatsResponse
	doJSON(
		t,
		http.MethodGet,
		"/stats/reviewers",
		nil,
		http.StatusOK,
		&statsResp,
	)

	if len(statsResp.Stats) == 0 {
		t.Fatalf("ожидали непустую статистику по ревьюверам")
	}

	expected := map[string]struct{}{
		"Reviewer1": {},
		"Reviewer2": {},
		"Reviewer3": {},
	}

	foundAny := false
	for _, item := range statsResp.Stats {
		if item.ReviewsCount < 0 {
			t.Fatalf(
				"reviews_count не может быть отрицательным, username=%s count=%d",
				item.Username,
				item.ReviewsCount,
			)
		}
		if _, ok := expected[item.Username]; ok && item.ReviewsCount > 0 {
			foundAny = true
		}
	}

	if !foundAny {
		t.Fatalf("ожидали, что хотя бы один из наших ревьюверов попадет в статистику с reviews_count > 0")
	}
}
