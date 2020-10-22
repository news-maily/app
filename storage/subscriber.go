package storage

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/mailbadger/app/entities"
)

// GetSubscribers fetches subscribers by user id, and populates the pagination obj
func (db *store) GetSubscribers(userID int64, p *PaginationCursor) error {
	p.SetCollection(&[]entities.Subscriber{})
	p.SetResource("subscribers")

	query := db.Table(p.Resource).
		Where("user_id = ?", userID).
		Order("created_at desc, id desc").
		Limit(p.PerPage)

	p.SetQuery(query)

	return db.Paginate(p, userID)
}

// GetSubscribersBySegmentID fetches subscribers by user id and list id, and populates the pagination obj
func (db *store) GetSubscribersBySegmentID(segmentID, userID int64, p *PaginationCursor) error {
	p.SetCollection(&[]entities.Subscriber{})
	p.SetResource("subscribers")
	p.SetScopes(BelongsToUser(userID), BelongsToSegment(segmentID))

	query := db.Table(p.Resource).
		Order("created_at desc, id desc").
		Limit(p.PerPage)

	p.SetQuery(query)

	return db.Paginate(p, userID)
}

// BelongsToSegment is a query scope that finds all subscribers under a segment id.
func BelongsToSegment(segID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("INNER JOIN subscribers_segments ON subscribers_segments.subscriber_id = subscribers.id").
			Where("subscribers_segments.segment_id = ?", segID)
	}
}

// GetTotalSubscribers fetches the total count by user id.
func (db *store) GetTotalSubscribers(userID int64) (int64, error) {
	var count int64
	err := db.Model(entities.Subscriber{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetTotalSubscribersBySegment fetches the total count by user and segment id.
func (db *store) GetTotalSubscribersBySegment(segmentID, userID int64) (int64, error) {
	var seg = entities.Segment{Model: entities.Model{ID: segmentID}}

	assoc := db.Model(&seg).Where("user_id = ?", userID).Association("Subscribers")
	return int64(assoc.Count()), assoc.Error
}

// GetSubscriber returns the subscriber by the given id and user id
func (db *store) GetSubscriber(id, userID int64) (*entities.Subscriber, error) {
	var s = new(entities.Subscriber)
	err := db.Preload("Segments").Where("user_id = ? and id = ?", userID, id).Find(s).Error
	return s, err
}

// GetSubscribersByIDs returns the subscriber by the given id and user id
func (db *store) GetSubscribersByIDs(ids []int64, userID int64) ([]entities.Subscriber, error) {
	var s []entities.Subscriber
	err := db.Where("user_id = ? and id in (?)", userID, ids).Find(&s).Error
	return s, err
}

// GetSubscriberByEmail returns the subscriber by the given email and user id
func (db *store) GetSubscriberByEmail(email string, userID int64) (*entities.Subscriber, error) {
	var s = new(entities.Subscriber)
	err := db.Where("user_id = ? and email = ?", userID, email).Find(s).Error
	return s, err
}

// GetDistinctSubscribersBySegmentIDs fetches all distinct subscribers by user id and list ids
func (db *store) GetDistinctSubscribersBySegmentIDs(
	listIDs []int64,
	userID int64,
	blacklisted, active bool,
	timestamp time.Time,
	nextID int64,
	limit int64,
) ([]entities.Subscriber, error) {
	if limit == 0 {
		limit = 1000
	}

	var subs []entities.Subscriber

	err := db.Table("subscribers").
		Select("DISTINCT(id), name, email, created_at").
		Joins("INNER JOIN subscribers_segments ON subscribers_segments.subscriber_id = subscribers.id").
		Where(`
			subscribers_segments.segment_id IN (?)
			AND subscribers.user_id = ? 
			AND subscribers.blacklisted = ? 
			AND subscribers.active = ?
			AND (created_at > ? OR (created_at = ? AND id > ?))
			AND created_at < ?`,
			listIDs,
			userID,
			blacklisted,
			active,
			timestamp.Format(time.RFC3339Nano),
			timestamp.Format(time.RFC3339Nano),
			nextID,
			time.Now().Format(time.RFC3339Nano)).
		Order("created_at, id").
		Limit(limit).
		Find(&subs).Error

	return subs, err
}

// CreateSubscriber creates a new subscriber in the database.
func (db *store) CreateSubscriber(s *entities.Subscriber) error {
	return db.Create(s).Error
}

// UpdateSubscriber edits an existing subscriber in the database.
func (db *store) UpdateSubscriber(s *entities.Subscriber) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(s).Association("Segments").Replace(s.Segments).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ? and user_id = ?", s.ID, s.UserID).Save(s).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeactivateSubscriber de-activates a subscriber by the given user and email.
func (db *store) DeactivateSubscriber(userID int64, email string) error {
	return db.Model(&entities.Subscriber{}).
		Where("user_id = ? AND email = ?", userID, email).
		Update("active", false).Error
}

// DeleteSubscriber deletes an existing subscriber from the database along with all his metadata.
func (db *store) DeleteSubscriber(id, userID int64) error {
	s, err := db.GetSubscriber(id, userID)
	if err != nil {
		return err
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(s).Association("Segments").Clear().Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("user_id = ?", userID).Delete(s).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteSubscriberByEmail deletes an existing subscriber by email from the database along with all his metadata.
func (db *store) DeleteSubscriberByEmail(email string, userID int64) error {
	s, err := db.GetSubscriberByEmail(email, userID)
	if err != nil {
		return err
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(s).Association("Segments").Clear().Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("user_id = ?", userID).Delete(s).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
