package storage

import (
	"testing"

	"github.com/news-maily/api/entities"
	"github.com/news-maily/api/utils/pagination"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	db := openTestDb()
	defer db.Close()

	store := From(db)

	//Test create list
	l := &entities.List{
		Name:   "foo",
		UserId: 1,
	}

	err := store.CreateList(l)
	assert.Nil(t, err)

	//Test get list
	l, err = store.GetList(l.Id, 1)
	assert.Nil(t, err)
	assert.Equal(t, l.Name, "foo")

	//Test update list
	l.Name = "bar"
	err = store.UpdateList(l)
	assert.Nil(t, err)
	assert.Equal(t, l.Name, "bar")

	//Test get list by name
	l, err = store.GetListByName("bar", 1)
	assert.Nil(t, err)
	assert.Equal(t, l.Name, "bar")

	//Test list validation when name is invalid
	l.Name = ""
	l.Validate()
	assert.Equal(t, l.Errors["name"], entities.ErrListNameEmpty.Error())

	//Test get lists
	p := &pagination.Pagination{PerPage: 10}
	store.GetLists(1, p)
	assert.NotEmpty(t, p.Collection)
	assert.Equal(t, len(p.Collection), int(p.Total))

	//Test append subscribers to list
	s := &entities.Subscriber{
		Name:   "john",
		Email:  "john@example.com",
		UserId: 1,
	}
	store.CreateSubscriber(s)

	l.Subscribers = append(l.Subscribers, *s)

	err = store.AppendSubscribers(l)
	assert.Nil(t, err)

	l, err = store.GetList(l.Id, 1)
	assert.NotEmpty(t, l.Subscribers)
	assert.Equal(t, l.Subscribers[0].Name, "john")

	//Test detach subscribers from list
	err = store.DetachSubscribers(l)
	assert.Nil(t, err)

	// Test delete list
	err = store.DeleteList(1, 1)
	assert.Nil(t, err)
}
