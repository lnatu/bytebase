package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bytebase/bytebase"
	"github.com/bytebase/bytebase/api"
	"github.com/bytebase/bytebase/bin/bb/connect"
	"github.com/bytebase/bytebase/bin/bb/dump/mysqldump"
	"github.com/google/jsonapi"
	"github.com/labstack/echo/v4"
)

func (s *Server) registerDatabaseRoutes(g *echo.Group) {
	g.POST("/database", func(c echo.Context) error {
		databaseCreate := &api.DatabaseCreate{}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, databaseCreate); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformatted create database request").SetInternal(err)
		}

		databaseCreate.CreatorId = c.Get(GetPrincipalIdContextKey()).(int)

		database, err := s.DatabaseService.CreateDatabase(context.Background(), databaseCreate)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ECONFLICT {
				return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Database name already exists: %s", databaseCreate.Name))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create database").SetInternal(err)
		}

		if err := s.ComposeDatabaseRelationship(context.Background(), database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created database relationship").SetInternal(err)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal create database response").SetInternal(err)
		}
		return nil
	})

	g.GET("/database", func(c echo.Context) error {
		databaseFind := &api.DatabaseFind{}
		if instanceIdStr := c.QueryParam("instance"); instanceIdStr != "" {
			instanceId, err := strconv.Atoi(instanceIdStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query parameter instance is not a number: %s", instanceIdStr)).SetInternal(err)
			}
			databaseFind.InstanceId = &instanceId
		}
		projectIdStr := c.QueryParams().Get("project")
		if projectIdStr != "" {
			projectId, err := strconv.Atoi(projectIdStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("project query parameter is not a number: %s", projectIdStr)).SetInternal(err)
			}
			databaseFind.ProjectId = &projectId
		}
		list, err := s.ComposeDatabaseListByFind(context.Background(), databaseFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch database list").SetInternal(err)
		}

		filteredList := []*api.Database{}
		role := c.Get(GetRoleContextKey()).(api.Role)
		// If caller is NOT requesting for a paritcular instance or the caller is a Developer,
		// then we will only return databases belonging to the project where the caller is a member of.
		// Looking from the UI perspective:
		// - The database list left sidebar will only return databases related to the caller regardless of the caller's role.
		// - The database list on the instance page will return all databases if the caller is Owner or DBA, but will only return
		//   related databases if the caller is Developer.
		if databaseFind.InstanceId == nil || role == api.Developer {
			principalId := c.Get(GetPrincipalIdContextKey()).(int)
			for _, database := range list {
				for _, projectMember := range database.Project.ProjectMemberList {
					if projectMember.PrincipalId == principalId {
						filteredList = append(filteredList, database)
						break
					}
				}
			}
		} else {
			filteredList = list
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, filteredList); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal database list response").SetInternal(err)
		}
		return nil
	})

	g.GET("/database/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		databaseFind := &api.DatabaseFind{
			ID: &id,
		}
		database, err := s.ComposeDatabaseByFind(context.Background(), databaseFind)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch database ID: %v", id)).SetInternal(err)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal project ID response: %v", id)).SetInternal(err)
		}
		return nil
	})

	g.PATCH("/database/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Database ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		databasePatch := &api.DatabasePatch{
			ID:        id,
			UpdaterId: c.Get(GetPrincipalIdContextKey()).(int),
		}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, databasePatch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformatted patch database request").SetInternal(err)
		}

		database, err := s.DatabaseService.PatchDatabase(context.Background(), databasePatch)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to patch database ID: %v", id)).SetInternal(err)
		}

		if err := s.ComposeDatabaseRelationship(context.Background(), database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch updated database relationship: %v", database.Name)).SetInternal(err)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, database); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal database ID response: %v", id)).SetInternal(err)
		}
		return nil
	})

	g.GET("/database/:id/table", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		databaseFind := &api.DatabaseFind{
			ID: &id,
		}
		database, err := s.ComposeDatabaseByFind(context.Background(), databaseFind)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch database ID: %v", id)).SetInternal(err)
		}

		tableFind := &api.TableFind{
			DatabaseId: &id,
		}
		tableList, err := s.TableService.FindTableList(context.Background(), tableFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch table list for database id: %d", id)).SetInternal(err)
		}

		for _, table := range tableList {
			table.Database = database
			columnFind := &api.ColumnFind{
				DatabaseId: &id,
				TableId:    &table.ID,
			}
			table.ColumnList, err = s.ColumnService.FindColumnList(context.Background(), columnFind)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch colmun list for database id: %d, table name: %s", id, table.Name)).SetInternal(err)
			}

			indexFind := &api.IndexFind{
				DatabaseId: &id,
				TableId:    &table.ID,
			}
			table.IndexList, err = s.IndexService.FindIndexList(context.Background(), indexFind)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch index list for database id: %d, table name: %s", id, table.Name)).SetInternal(err)
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, tableList); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal fetch table list response: %v", id)).SetInternal(err)
		}
		return nil
	})

	g.GET("/database/:id/table/:tableName", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		databaseFind := &api.DatabaseFind{
			ID: &id,
		}
		database, err := s.ComposeDatabaseByFind(context.Background(), databaseFind)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch database ID: %v", id)).SetInternal(err)
		}

		tableName := c.Param("tableName")
		tableFind := &api.TableFind{
			DatabaseId: &id,
			Name:       &tableName,
		}
		table, err := s.TableService.FindTable(context.Background(), tableFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch table for database id: %d, table name: %s", id, tableName)).SetInternal(err)
		}

		table.Database = database

		columnFind := &api.ColumnFind{
			DatabaseId: &id,
			TableId:    &table.ID,
		}
		table.ColumnList, err = s.ColumnService.FindColumnList(context.Background(), columnFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch colmun list for database id: %d, table name: %s", id, tableName)).SetInternal(err)
		}

		indexFind := &api.IndexFind{
			DatabaseId: &id,
			TableId:    &table.ID,
		}
		table.IndexList, err = s.IndexService.FindIndexList(context.Background(), indexFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch index list for database id: %d, table name: %s", id, table.Name)).SetInternal(err)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, table); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal fetch table response: %v", id)).SetInternal(err)
		}
		return nil
	})

	g.POST("/database/:id/backup", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		backupCreate := &api.BackupCreate{}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, backupCreate); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformatted create database request").SetInternal(err)
		}
		backupCreate.CreatorId = c.Get(GetPrincipalIdContextKey()).(int)

		databaseFind := &api.DatabaseFind{
			ID: &id,
		}
		database, err := s.ComposeDatabaseByFind(context.Background(), databaseFind)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch database ID: %v", id)).SetInternal(err)
		}

		backup, err := s.BackupService.CreateBackup(context.Background(), backupCreate)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ECONFLICT {
				return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Backup name already exists: %s", backupCreate.Name))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create backup").SetInternal(err)
		}
		backup.Database = database

		if err := s.BackupDatabase(backup); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create backup %q.", backup.Name)).SetInternal(err)
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, backup); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal create backup response").SetInternal(err)
		}
		return nil
	})

	g.GET("/database/:id/backup", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("ID is not a number: %s", c.Param("id"))).SetInternal(err)
		}

		databaseFind := &api.DatabaseFind{
			ID: &id,
		}
		database, err := s.ComposeDatabaseByFind(context.Background(), databaseFind)
		if err != nil {
			if bytebase.ErrorCode(err) == bytebase.ENOTFOUND {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Database ID not found: %d", id))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch database ID: %v", id)).SetInternal(err)
		}

		backupFind := &api.BackupFind{
			DatabaseId: &id,
		}
		backupList, err := s.BackupService.FindBackupList(context.Background(), backupFind)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to backup list for database id: %d", id)).SetInternal(err)
		}

		for _, backup := range backupList {
			backup.Database = database
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, backupList); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal fetch table list response: %v", id)).SetInternal(err)
		}
		return nil
	})
}

func (s *Server) ComposeDatabaseByFind(ctx context.Context, find *api.DatabaseFind) (*api.Database, error) {
	database, err := s.DatabaseService.FindDatabase(context.Background(), find)
	if err != nil {
		return nil, err
	}

	if err := s.ComposeDatabaseRelationship(ctx, database); err != nil {
		return nil, err
	}

	return database, nil
}

func (s *Server) ComposeDatabaseListByFind(ctx context.Context, find *api.DatabaseFind) ([]*api.Database, error) {
	list, err := s.DatabaseService.FindDatabaseList(context.Background(), find)
	if err != nil {
		return nil, err
	}

	for _, database := range list {
		if err := s.ComposeDatabaseRelationship(context.Background(), database); err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (s *Server) ComposeDatabaseRelationship(ctx context.Context, database *api.Database) error {
	var err error

	database.Creator, err = s.ComposePrincipalById(context.Background(), database.CreatorId)
	if err != nil {
		return err
	}

	database.Updater, err = s.ComposePrincipalById(context.Background(), database.UpdaterId)
	if err != nil {
		return err
	}

	database.Project, err = s.ComposeProjectlById(context.Background(), database.ProjectId)
	if err != nil {
		return err
	}

	database.Instance, err = s.ComposeInstanceById(context.Background(), database.InstanceId)
	if err != nil {
		return err
	}

	database.DataSourceList = []*api.DataSource{}

	return nil
}

// BackupDatabase will take a backup of a database.
func (s *Server) BackupDatabase(backup *api.Backup) error {
	instance := backup.Database.Instance
	conn, err := connect.NewMysql(instance.Username, instance.Password, instance.Host, instance.Port, backup.Database.Name, nil /* tlsConfig */)
	if err != nil {
		return fmt.Errorf("connect.NewMysql(%q, %q, %q, %q) got error: %v", instance.Username, instance.Password, instance.Host, instance.Port, err)
	}
	defer conn.Close()
	dp := mysqldump.New(conn)

	f, err := os.Create(backup.Path)
	if err != nil {
		return fmt.Errorf("failed to open backup path: %s", backup.Path)
	}
	defer f.Close()

	if err := dp.Dump(backup.Database.Name, f, false /* schemaOnly */); err != nil {
		return err
	}

	return nil
}
