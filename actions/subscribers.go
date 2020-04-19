package actions

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/news-maily/app/entities"
	"github.com/news-maily/app/logger"
	"github.com/news-maily/app/routes/middleware"
	"github.com/news-maily/app/storage"
	"github.com/sirupsen/logrus"
)

func GetSubscribers(c *gin.Context) {
	val, ok := c.Get("cursor")
	if !ok {
		logger.From(c).Error("Unable to fetch pagination cursor from context.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to fetch segments. Please try again.",
		})
		return
	}

	p, ok := val.(*storage.PaginationCursor)
	if !ok {
		logger.From(c).Error("Unable to cast pagination cursor from context value.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to fetch segments. Please try again.",
		})
		return
	}

	err := storage.GetSubscribers(c, middleware.GetUser(c).ID, p)
	if err != nil {
		logger.From(c).WithError(err).Error("Unable to fetch subscribers collection.")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to fetch subscribers. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, p)
}

func GetSubscriber(c *gin.Context) {
	if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
		if s, err := storage.GetSubscriber(c, id, middleware.GetUser(c).ID); err == nil {
			c.JSON(http.StatusOK, s)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{
			"message": "Subscriber not found",
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"message": "Id must be an integer",
	})
}

type segmentsParam struct {
	Ids []int64 `form:"segments[]"`
}

func PostSubscriber(c *gin.Context) {
	name, email := strings.TrimSpace(c.PostForm("name")), strings.TrimSpace(c.PostForm("email"))
	meta := c.PostFormMap("metadata")
	segments := &segmentsParam{}
	err := c.Bind(segments)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Invalid data",
			"errors": map[string]string{
				"segments": "The segments array is in an invalid format.",
			},
		})
		return
	}

	s := &entities.Subscriber{
		Name:     name,
		Email:    email,
		Metadata: meta,
		UserID:   middleware.GetUser(c).ID,
	}

	segs, err := storage.GetSegmentsByIDs(c, s.UserID, segments.Ids)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Invalid data",
			"errors": map[string]string{
				"segments": "Unable to find the specified segments.",
			},
		})
		return
	}

	s.Segments = segs

	if !s.Validate() {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Invalid data",
			"errors":  s.Errors,
		})
		return
	}

	_, err = storage.GetSubscriberByEmail(c, s.Email, s.UserID)
	if err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Subscriber with that email already exists.",
		})
		return
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Unable to create subscriber, invalid metadata.",
		})
		return
	}
	s.MetaJSON = metaJSON

	if err := storage.CreateSubscriber(c, s); err != nil {
		logger.From(c).
			WithError(err).
			Warn("Unable to create subscriber.")
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Unable to create subscriber",
		})
		return
	}

	c.JSON(http.StatusCreated, s)
}

func PutSubscriber(c *gin.Context) {
	if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
		s, err := storage.GetSubscriber(c, id, middleware.GetUser(c).ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Subscriber not found",
			})
			return
		}

		s.Name = strings.TrimSpace(c.PostForm("name"))
		s.Email = strings.TrimSpace(c.PostForm("email"))

		if !s.Validate() {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Invalid data",
				"errors":  s.Errors,
			})
			return
		}

		if err = storage.UpdateSubscriber(c, s); err != nil {
			logger.From(c).
				WithError(err).
				WithField("subscriber_id", id).
				Warn("Unable to update subscriber.")
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Unable to update subscriber.",
			})
			return
		}

		c.Status(http.StatusNoContent)
		return

	}

	c.JSON(http.StatusBadRequest, gin.H{
		"message": "Id must be an integer",
	})
}

func DeleteSubscriber(c *gin.Context) {
	if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
		user := middleware.GetUser(c)
		_, err := storage.GetSubscriber(c, id, user.ID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Subscriber not found",
			})
			return
		}

		err = storage.DeleteSubscriber(c, id, user.ID)
		if err != nil {
			logger.From(c).WithError(err).Warn("Unable to delete subscriber.")
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "Unable to delete subscriber.",
			})
			return
		}

		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"message": "Id must be an integer",
	})
}

func PostUnsubscribe(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	uuid := strings.TrimSpace(c.PostForm("uuid"))
	token := strings.TrimSpace(c.PostForm("t"))

	if token == "" || email == "" || uuid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to process the request, invalid data sent.",
		})
		return
	}

	u, err := storage.GetUserByUUID(c, uuid)
	if err != nil {
		logger.From(c).WithFields(logrus.Fields{
			"email": email,
			"uuid":  uuid,
		}).WithError(err).Warn("Unsubscribe: cannot find user by uuid.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to process the request, invalid data sent.",
		})
		return
	}

	sub, err := storage.GetSubscriberByEmail(c, email, u.ID)
	if err != nil {
		logger.From(c).WithFields(logrus.Fields{
			"email": email,
			"uuid":  uuid,
		}).WithError(err).Warn("Unsubscribe: unable to fetch subscriber by email.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to process the request, invalid data sent.",
		})
		return
	}

	hash, err := sub.GenerateUnsubscribeToken(os.Getenv("UNSUBSCRIBE_SECRET"))
	if err != nil {
		logger.From(c).WithFields(logrus.Fields{
			"email": email,
			"uuid":  uuid,
		}).WithError(err).Error("Unsubscribe: unable to generate hash.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to process the request, invalid data sent.",
		})
		return
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(hash)) != 1 {
		logger.From(c).WithFields(logrus.Fields{
			"email": email,
			"uuid":  uuid,
		}).WithError(err).Warn("Unsubscribe: hashes don't match.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to process the request, invalid data sent.",
		})
		return
	}

	sub.Active = false
	err = storage.UpdateSubscriber(c, sub)
	if err != nil {
		logger.From(c).WithFields(logrus.Fields{
			"email": email,
			"uuid":  uuid,
		}).WithError(err).Warn("Unsubscribe: unable to update subscriber's status.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to process the request, please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "You have successfully unsusbcribed.",
	})
}
