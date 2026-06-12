package oauthrepo

import "errors"

var (
	ErrTransactionRollbackFailed = errors.New("transaction rollback failed")
	ErrTransactionCommitFailed   = errors.New("transaction commit failed")
)
