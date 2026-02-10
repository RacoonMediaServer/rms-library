package db

import (
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go.mongodb.org/mongo-driver/bson"
)

func getSort(sort *rms_library.Sort) bson.D {
	sortField := "title"
	sortOrder := 1
	if sort != nil {
		switch sort.By {
		case rms_library.Sort_Title:
		case rms_library.Sort_CreatedAt:
			sortField = "createdat"
		}
		if sort.Order == rms_library.Sort_Desc {
			sortOrder = -1
		}
	}

	return bson.D{{sortField, sortOrder}}
}
