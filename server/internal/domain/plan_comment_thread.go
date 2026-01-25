package domain

import "time"

type PlanCommentThreadStatus string

const (
	PlanCommentThreadStatusActive   PlanCommentThreadStatus = "active"
	PlanCommentThreadStatusResolved PlanCommentThreadStatus = "resolved"
	PlanCommentThreadStatusOutdated PlanCommentThreadStatus = "outdated" // 本文変更で無効化
)

func (s PlanCommentThreadStatus) IsValid() bool {
	switch s {
	case PlanCommentThreadStatusActive, PlanCommentThreadStatusResolved, PlanCommentThreadStatusOutdated:
		return true
	}
	return false
}

// PlanCommentThread represents a comment thread anchored to a specific location in a PlanDocument
type PlanCommentThread struct {
	ID             string
	PlanDocumentID string // reference to PlanDocument

	// 位置特定情報
	TargetText       string // コメント対象のテキスト（選択範囲）
	ContextBefore    string // 対象テキストの前のコンテキスト（〜100文字）
	ContextAfter     string // 対象テキストの後のコンテキスト（〜100文字）
	OriginalBodyHash string // コメント作成時のPlan本文のハッシュ

	Status PlanCommentThreadStatus // active / resolved / outdated

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CommentPosition represents the calculated position of a comment thread in the current document body
type CommentPosition struct {
	StartOffset int  // 開始位置（文字オフセット）
	EndOffset   int  // 終了位置（文字オフセット）
	StartLine   int  // 開始行（1-indexed）
	StartColumn int  // 開始列（1-indexed）
	EndLine     int  // 終了行（1-indexed）
	EndColumn   int  // 終了列（1-indexed）
	Found       bool // 位置が見つかったかどうか
}
