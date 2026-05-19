// Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.
// Project repository: https://github.com/zhian9/go-admin

package biz

import (
	"context"

	"github.com/zhian9/blogo-server/internal/mods/blog/dal"
	bschema "github.com/zhian9/blogo-server/internal/mods/blog/schema"
	rdal "github.com/zhian9/blogo-server/internal/mods/rbac/dal"
	rschema "github.com/zhian9/blogo-server/internal/mods/rbac/schema"
	"github.com/zhian9/blogo-server/pkg/util"
)

// ProfileDashboard 个人主页聚合查询结果
type ProfileDashboard struct {
	User              *rschema.User              `json:"user"`
	Articles          *bschema.Articles          `json:"articles"`
	LikedArticles     bschema.Articles           `json:"liked_articles"`
	FavArticles       bschema.Articles           `json:"favorite_articles"`
	Comments          bschema.Comments           `json:"comments"`
	Followers         []string                   `json:"followers"`
	Following         []string                   `json:"following"`
	Contributions     []*bschema.ContributionDay `json:"contributions"`
	ContributionStats *bschema.ContributionStats `json:"contribution_stats"`
}

// Profile 个人主页业务（跨模块聚合）
type Profile struct {
	ArticleDAL         *dal.Article
	ArticleLikeDAL     *dal.ArticleLike
	ArticleFavoriteDAL *dal.ArticleFavorite
	CommentDAL         *dal.Comment
	ContributionDAL    *dal.Contribution
	UserDAL            *rdal.User
	UserFollowDAL      *rdal.UserFollow
}

// GetDashboard 聚合查询个人主页数据。
// viewerID: 当前查看者的用户ID（空=未登录）
// targetUserID: 要查看的用户ID
func (p *Profile) GetDashboard(ctx context.Context, viewerID, targetUserID string) (*ProfileDashboard, error) {
	dash := &ProfileDashboard{}

	// 1. 用户基本信息
	user, err := p.UserDAL.Get(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	dash.User = user

	// 2. 文章列表：仅展示 targetUserID 的文章
	isSelf := viewerID != "" && viewerID == targetUserID
	if isSelf {
		result, err := p.ArticleDAL.Query(ctx, bschema.ArticleQueryParam{
			PaginationParam: util.PaginationParam{Current: 1, PageSize: 20},
			AuthorID:        targetUserID,
		}, bschema.ArticleQueryOptions{
			UserID:         viewerID,
			WithCategory:   true,
			WithTags:       true,
			WithCoverImage: true,
			QueryOptions: util.QueryOptions{
				OrderFields: []util.OrderByParam{
					{Field: "created_at", Direction: util.DESC},
				},
			},
		})
		if err == nil {
			dash.Articles = &result.Data
		}
	} else {
		result, err := p.ArticleDAL.Query(ctx, bschema.ArticleQueryParam{
			PaginationParam: util.PaginationParam{Current: 1, PageSize: 20},
			AuthorID:        targetUserID,
			Status:          bschema.ArticleStatusPublished,
		}, bschema.ArticleQueryOptions{
			WithCategory:   true,
			WithTags:       true,
			WithCoverImage: true,
			QueryOptions: util.QueryOptions{
				OrderFields: []util.OrderByParam{
					{Field: "created_at", Direction: util.DESC},
				},
			},
		})
		if err == nil {
			dash.Articles = &result.Data
		}
	}

	// 3. 点赞的文章（仅公开已发布）
	likedIDs, _ := p.ArticleLikeDAL.GetLikedArticleIDs(ctx, targetUserID)
	if len(likedIDs) > 0 {
		likedResult, _ := p.ArticleDAL.Query(ctx, bschema.ArticleQueryParam{
			PaginationParam: util.PaginationParam{Current: 1, PageSize: 20},
			Status:          bschema.ArticleStatusPublished,
		}, bschema.ArticleQueryOptions{
			WithCategory:   true,
			WithTags:       true,
			WithCoverImage: true,
		})
		if likedResult != nil {
			var likedArticles bschema.Articles
			for _, a := range likedResult.Data {
				for _, id := range likedIDs {
					if a.ID == id {
						likedArticles = append(likedArticles, a)
						break
					}
				}
			}
			dash.LikedArticles = likedArticles
		}
	}

	// 4. 收藏的文章（仅公开已发布）
	favIDs, _ := p.ArticleFavoriteDAL.GetArticleIDsByUserID(ctx, targetUserID, 20)
	if len(favIDs) > 0 {
		favResult, _ := p.ArticleDAL.Query(ctx, bschema.ArticleQueryParam{
			PaginationParam: util.PaginationParam{Current: 1, PageSize: 20},
			Status:          bschema.ArticleStatusPublished,
		}, bschema.ArticleQueryOptions{
			WithCategory:   true,
			WithTags:       true,
			WithCoverImage: true,
		})
		if favResult != nil {
			var favArticles bschema.Articles
			for _, a := range favResult.Data {
				for _, id := range favIDs {
					if a.ID == id {
						favArticles = append(favArticles, a)
						break
					}
				}
			}
			dash.FavArticles = favArticles
		}
	}

	// 5. 评论记录
	commentResult, _ := p.CommentDAL.Query(ctx, bschema.CommentQueryParam{
		UserID:          targetUserID,
		Status:          bschema.CommentStatusApproved,
		PaginationParam: util.PaginationParam{Current: 1, PageSize: 20},
	}, bschema.CommentQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{{Field: "created_at", Direction: util.DESC}},
		},
	})
	if commentResult != nil {
		dash.Comments = commentResult.Data
	}

	// 6. 粉丝与关注列表
	followers, _, _ := p.UserFollowDAL.ListFollowers(ctx, targetUserID, util.PaginationParam{Current: 1, PageSize: 50})
	following, _, _ := p.UserFollowDAL.ListFollowing(ctx, targetUserID, util.PaginationParam{Current: 1, PageSize: 50})
	dash.Followers = followers
	dash.Following = following
	if dash.Followers == nil {
		dash.Followers = []string{}
	}
	if dash.Following == nil {
		dash.Following = []string{}
	}

	// 7. 贡献活跃数据
	contributions, err := p.ContributionDAL.ComputeContributions(ctx, targetUserID)
	if err != nil {
		contributions = []*bschema.ContributionDay{}
	}
	dash.Contributions = contributions
	dash.ContributionStats = p.ContributionDAL.ComputeStats(contributions)

	return dash, nil
}
