package model

import (
	"strings"

	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

var contentTypePrefix = map[rms_library.ContentType]string{
	rms_library.ContentType_TypeMovies: "mov:",
	rms_library.ContentType_TypeMusic:  "mus:",
	rms_library.ContentType_TypeOther:  "other:",
}

type ID string

func (id ID) ContentType() rms_library.ContentType {
	for contentType, prefix := range contentTypePrefix {
		if strings.HasPrefix(string(id), prefix) {
			return contentType
		}
	}
	return rms_library.ContentType_TypeMovies
}

func MakeID(id string, contentType rms_library.ContentType) ID {
	return ID(id + contentTypePrefix[contentType])
}

func (id ID) String() string {
	return string(id)
}
