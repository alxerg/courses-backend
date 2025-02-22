package repository

import "errors"

var (
	ErrUserNotFound            = errors.New("user doesn't exists")
	ErrVerificationCodeInvalid = errors.New("verification code is invalid")
	ErrOfferNotFound           = errors.New("offer doesn't exists")
	ErrPromoNotFound           = errors.New("promocode doesn't exists")
)
