package storage

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mailbadger/app/entities"
	"github.com/stretchr/testify/assert"
)

func TestDeliveries(t *testing.T) {
	db := openTestDb()
	defer func() {
		err := db.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

	store := From(db)
	now := time.Now().UTC()

	// Test insert deliveries
	deliveries := []entities.Delivery{
		{
			ID:                   1,
			UserID:               1,
			CampaignID:           1,
			Recipient:            "jhon@email.com",
			ProcessingTimeMillis: 0,
			SMTPResponse:         "a",
			ReportingMTA:         "a",
			RemoteMtaIP:          "a",
			CreatedAt:            now,
		},
		{
			ID:                   2,
			UserID:               1,
			CampaignID:           1,
			Recipient:            "jhon2@email.com",
			ProcessingTimeMillis: 0,
			SMTPResponse:         "a",
			ReportingMTA:         "a",
			RemoteMtaIP:          "a",
			CreatedAt:            now,
		},
	}
	// insert delivery 1
	err := store.CreateDelivery(&deliveries[0])
	assert.Nil(t, err)
	// insert delivery 2
	err = store.CreateDelivery(&deliveries[1])
	assert.Nil(t, err)

	// test get total delivery stats
	totalDeliveries, err := store.GetTotalDelivered(1)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), totalDeliveries)

}
