package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// handleGetSegments retrieves segments with additional metadata like subscriber counts. This may be slow.
func handleGetSegments(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
		pg  = app.paginator.NewFromURL(c.Request().URL.Query())

		query        = strings.TrimSpace(c.FormValue("query"))
		orderBy      = c.FormValue("order_by")
		order        = c.FormValue("order")
		minimal, _   = strconv.ParseBool(c.FormValue("minimal"))
		segmentID, _ = strconv.Atoi(c.Param("id"))

		out models.PageResults
	)

	// Fetch one segment.
	single := false
	if segmentID > 0 {
		single = true
	}

	if single {
		out, err := app.core.GetSegment(segmentID, "")
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, okResp{out})
	}

	// Minimal query simply returns the segment of all segments without JOIN subscriber counts. This is fast.
	if !single && minimal {
		res, err := app.core.GetSegments()
		if err != nil {
			return err
		}
		if len(res) == 0 {
			return c.JSON(http.StatusOK, okResp{[]struct{}{}})
		}

		// Meta.
		out.Results = res
		out.Total = len(res)
		out.Page = 1
		out.PerPage = out.Total

		return c.JSON(http.StatusOK, okResp{out})
	}

	// Full segment query.
	res, total, err := app.core.QuerySegments(query, orderBy, order, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	if single && len(res) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			app.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.segment}"))
	}

	if single {
		return c.JSON(http.StatusOK, okResp{res[0]})
	}

	out.Query = query
	out.Results = res
	out.Total = total
	out.Page = pg.Page
	out.PerPage = pg.PerPage

	return c.JSON(http.StatusOK, okResp{out})
}

// handleCreateSegment handles segment creation.
func handleCreateSegment(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
		l   = models.Segment{}
	)

	if err := c.Bind(&l); err != nil {
		return err
	}

	// Validate.
	if !strHasLen(l.Name, 1, stdInputMaxLen) {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("segments.invalidName"))
	}

	out, err := app.core.CreateSegment(l)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{out})
}

// handleUpdateSegment handles segment modification.
func handleUpdateSegment(c echo.Context) error {
	var (
		app   = c.Get("app").(*App)
		id, _ = strconv.Atoi(c.Param("id"))
	)

	if id < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidID"))
	}

	// Incoming params.
	var l models.Segment
	if err := c.Bind(&l); err != nil {
		return err
	}

	// Validate.
	if !strHasLen(l.Name, 1, stdInputMaxLen) {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("segments.invalidName"))
	}

	out, err := app.core.UpdateSegment(id, l)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{out})
}

// handleDeleteSegments handles segment deletion, either a single one (ID in the URI), or a segment.
func handleDeleteSegments(c echo.Context) error {
	var (
		app   = c.Get("app").(*App)
		id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
		ids   []int
	)

	if id < 1 && len(ids) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidID"))
	}

	if id > 0 {
		ids = append(ids, int(id))
	}

	if err := app.core.DeleteSegments(ids); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{true})
}

// handleCountSubscribersByQuery handles validate and count subscribers in segment.
func handleCountSubscribersByQuery(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
	)

	// Incoming params.
	var l models.Segment
	if err := c.Bind(&l); err != nil {
		return err
	}

	out, err := app.core.CountSubscribersByQuery(l.SegmentQuery)
	if err != nil {
		return err
	}

	resp := map[string]interface{}{}
	resp["total"] = out
	return c.JSON(http.StatusOK, okResp{resp})
}
