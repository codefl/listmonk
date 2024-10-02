package core

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// GetSegments gets all segments optionally filtered by type.
func (c *Core) GetSegments() ([]models.Segment, error) {
	out := []models.Segment{}

	if err := c.q.GetSegments.Select(&out, "id"); err != nil {
		c.log.Printf("error fetching segments: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.segments}", "error", pqErrMsg(err)))
	}

	return out, nil
}

// QuerySegments gets multiple segments based on multiple query params. Along with the  paginated and sliced
// results, the total number of segments in the DB is returned.
func (c *Core) QuerySegments(searchStr, orderBy, order string, offset, limit int) ([]models.Segment, int, error) {
	var (
		out            = []models.Segment{}
		queryStr, stmt = makeSearchQuery(searchStr, orderBy, order, c.q.QuerySegments, []string{"name", "created_at", "updated_at"})
	)
	if err := c.db.Select(&out, stmt, 0, "", queryStr, offset, limit); err != nil {
		c.log.Printf("error fetching segments: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.segments}", "error", pqErrMsg(err)))
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// GetSegment gets a segment by its ID or UUID.
func (c *Core) GetSegment(id int, uuid string) (models.Segment, error) {
	var uu interface{}
	if uuid != "" {
		uu = uuid
	}

	var res []models.Segment
	queryStr, stmt := makeSearchQuery("", "", "", c.q.QuerySegments, nil)
	if err := c.db.Select(&res, stmt, id, uu, queryStr, 0, 1); err != nil {
		c.log.Printf("error fetching segments: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.segments}", "error", pqErrMsg(err)))
	}

	if len(res) == 0 {
		return models.Segment{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.segment}"))
	}

	out := res[0]

	return out, nil
}

// CreateSegment creates a new segment.
func (c *Core) CreateSegment(l models.Segment) (models.Segment, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		c.log.Printf("error generating UUID: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUUID", "error", err.Error()))
	}

	// Insert and read ID.
	var newID int
	l.UUID = uu.String()
	if err := c.q.CreateSegment.Get(&newID, l.UUID, l.Name, l.SegmentQuery, l.Description); err != nil {
		c.log.Printf("error creating segment: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.terms.segment}", "error", pqErrMsg(err)))
	}

	return c.GetSegment(newID, "")
}

// UpdateSegment updates a given segment.
func (c *Core) UpdateSegment(id int, l models.Segment) (models.Segment, error) {
	res, err := c.q.UpdateSegment.Exec(id, l.Name, l.SegmentQuery, l.Description)
	if err != nil {
		c.log.Printf("error updating segment: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.segment}", "error", pqErrMsg(err)))
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return models.Segment{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.segment}"))
	}

	return c.GetSegment(id, "")
}

// DeleteSegment deletes a segment.
func (c *Core) DeleteSegment(id int) error {
	return c.DeleteSegments([]int{id})
}

// DeleteSegments deletes multiple segments.
func (c *Core) DeleteSegments(ids []int) error {
	if _, err := c.q.DeleteSegments.Exec(pq.Array(ids)); err != nil {
		c.log.Printf("error deleting segments: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "{globals.terms.segment}", "error", pqErrMsg(err)))
	}
	return nil
}

// CountSubscribersInSegment count subscribers using segment query
func (c *Core) CountSubscribersByQuery(query string) (int, error) {
	// If there's no condition, it's a "get all" call which can probably be optionally pulled from cache.
	cond := query
	if query != "" {
		cond = fmt.Sprintf(" WHERE %s", query)
	}
	c.log.Printf(cond)

	// Create a readonly transaction that just does COUNT() to obtain the count of results
	// and to ensure that the arbitrary query is indeed readonly.
	stmt := fmt.Sprintf(c.q.CountSubscribersByQuery, cond)

	tx, err := c.db.BeginTxx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		c.log.Printf("error preparing subscriber query: %v", err)
		return 0, echo.NewHTTPError(http.StatusBadRequest, c.i18n.Ts("subscribers.errorPreparingQuery", "error", pqErrMsg(err)))
	}
	defer tx.Rollback()

	// Execute the readonly query and get the count of results.
	total := 0
	if err := tx.Get(&total, stmt); err != nil {
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.subscribers}", "error", pqErrMsg(err)))
	}

	return total, nil
}
