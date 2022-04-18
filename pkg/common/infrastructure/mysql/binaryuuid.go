package mysql

import (
	"database/sql/driver"

	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type BinaryUUID uuid.UUID

func (uid BinaryUUID) Value() (driver.Value, error) {
	return uuid.UUID(uid).Bytes(), nil
}

func (uid *BinaryUUID) Scan(src interface{}) error {
	var result uuid.UUID
	err := result.Scan(src)
	*uid = BinaryUUID(result)
	return err
}

func ConvertToUuids(iDs []uuid.UUID) []BinaryUUID {
	result := make([]BinaryUUID, 0, len(iDs))
	for _, id := range iDs {
		result = append(result, BinaryUUID(id))
	}
	return result
}
