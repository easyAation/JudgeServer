package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"online_judge/talcity/scaffold/criteria/merr"
	"online_judge/talcity/scaffold/util"

	"github.com/ttacon/libphonenumber"
)

const (
	EmailFormat = `^\w[-\w.+]*@([A-Za-z0-9][-A-Za-z0-9]+\.)+[A-Za-z]{2,14}$`
)

var (
	emptyStrErr             = errors.New("empty string")
	phoneFormatErr          = errors.New("phone format error, expect {+CC}{Space}{Number}")
	invalidCountryCodeErr   = errors.New("invalid country code")
	unknownCountryRegionErr = errors.New("unknown country code region")
	invalidPhoneErr         = errors.New("invalid phone number")
	invalidEmailErr         = errors.New("invalid email addr")
)

func IsEmail(email string) bool {
	if email == "" {
		return false
	}
	matched, _ := regexp.MatchString(EmailFormat, email)
	return matched
}

func ValidPhone(str string) error {
	if str == "" {
		return merr.WrapDefaultCode(emptyStrErr)
	}

	if str[0] != '+' {
		return merr.WrapDefaultCode(phoneFormatErr)
	}
	pieces := strings.Split(str, " ")
	if len(pieces) != 2 {
		return merr.WrapDefaultCode(phoneFormatErr)
	}

	countryCodeStr := pieces[0]
	countryCode, err := strconv.Atoi(strings.TrimLeft(countryCodeStr, "+"))
	if err != nil {
		return merr.WrapDefaultCode(invalidCountryCodeErr)
	}

	region := libphonenumber.GetRegionCodeForCountryCode(countryCode)
	if region == libphonenumber.UNKNOWN_REGION {
		return merr.WrapDefaultCode(unknownCountryRegionErr)
	}

	phoneNumber, err := libphonenumber.Parse(str, region)
	if err != nil {
		return merr.WrapDefaultCode(phoneFormatErr)
	}

	if !libphonenumber.IsValidNumber(phoneNumber) {
		return merr.WrapDefaultCode(phoneFormatErr)
	}

	return nil
}

func ValidatePhoneEmail(mediaType, phone, email string) error {
	if mediaType == util.MediaTypeEmail {
		if IsEmail(email) {
			return nil
		}

		return merr.WrapDefaultCode(invalidEmailErr)
	}
	if mediaType == util.MediaTypePhone {
		return ValidPhone(phone)
	}

	return merr.WrapDefaultCode(fmt.Errorf("unsupport media type %s", mediaType))
}
