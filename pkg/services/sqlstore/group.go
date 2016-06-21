package sqlstore

import (
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/events"
	m "github.com/grafana/grafana/pkg/models"
)

func init() {
	bus.AddHandler("sql", GetGroupById)
	bus.AddHandler("sql", CreateGroup)
	bus.AddHandler("sql", UpdateGroup)
	bus.AddHandler("sql", UpdateGroupAddress)
	bus.AddHandler("sql", GetGroupByName)
	bus.AddHandler("sql", SearchGroup)
	bus.AddHandler("sql", DeleteGroup)
}

func SearchGroups(query *m.SearchGroupsQuery) error {
	query.Result = make([]*m.GroupDTO, 0)
	sess := x.Table("group")
	if query.Query != "" {
		sess.Where("name LIKE ?", query.Query+"%")
	}
	if query.Name != "" {
		sess.Where("name=?", query.Name)
	}
	sess.Limit(query.Limit, query.Limit*query.Page)
	sess.Cols("id", "name")
	err := sess.Find(&query.Result)
	return err
}

func GetGroupById(query *m.GetGroupByIdQuery) error {
	var group m.Group
	exists, err := x.Id(query.Id).Get(&group)
	if err != nil {
		return err
	}

	if !exists {
		return m.ErrGroupNotFound
	}

	query.Result = &group
	return nil
}

func GetGroupByName(query *m.GetGroupByNameQuery) error {
	var group m.Group
	exists, err := x.Where("name=?", query.Name).Get(&group)
	if err != nil {
		return err
	}

	if !exists {
		return m.ErrGroupNotFound
	}

	query.Result = &group
	return nil
}

func isGroupNameTaken(name string, existingId int64, sess *session) (bool, error) {
	// check if org name is taken
	var group m.Group
	exists, err := sess.Where("name=?", name).Get(&group)

	if err != nil {
		return false, nil
	}

	if exists && existingId != group.Id {
		return true, nil
	}

	return false, nil
}

func CreateGroup(cmd *m.CreateGroupCommand) error {
	return inTransaction2(func(sess *session) error {

		if isNameTaken, err := isGroupNameTaken(cmd.Name, 0, sess); err != nil {
			return err
		} else if isNameTaken {
			return m.ErrGroupNameTaken
		}

		group := m.Group{
			Name:    cmd.Name,
			Created: time.Now(),
			Updated: time.Now(),
		}

		if _, err := sess.Insert(&group); err != nil {
			return err
		}

		user := m.GroupUser{
			GroupId:   group.Id,
			UserId:  cmd.UserId,
			Role:    m.ROLE_ADMIN,
			Created: time.Now(),
			Updated: time.Now(),
		}

		_, err := sess.Insert(&user)
		cmd.Result = group

		sess.publishAfterCommit(&events.GroupCreated{
			Timestamp: group.Created,
			Id:        group.Id,
			Name:      group.Name,
		})

		return err
	})
}

func UpdateGroup(cmd *m.UpdateGroupCommand) error {
	return inTransaction2(func(sess *session) error {

		if isNameTaken, err := isGroupNameTaken(cmd.Name, cmd.GroupId, sess); err != nil {
			return err
		} else if isNameTaken {
			return m.ErrGroupNameTaken
		}

		group := m.Group{
			Name:    cmd.Name,
			Updated: time.Now(),
		}

		if _, err := sess.Id(cmd.GroupId).Update(&group); err != nil {
			return err
		}

		sess.publishAfterCommit(&events.GroupUpdated{
			Timestamp: group.Updated,
			Id:        group.Id,
			Name:      group.Name,
		})

		return nil
	})
}

func DeleteGroup(cmd *m.DeleteGroupCommand) error {
	return inTransaction2(func(sess *session) error {

		deletes := []string{
			"DELETE FROM star WHERE EXISTS (SELECT 1 FROM dashboard WHERE group_id = ? AND star.dashboard_id = dashboard.id)",
			"DELETE FROM dashboard_tag WHERE EXISTS (SELECT 1 FROM dashboard WHERE group_id = ? AND dashboard_tag.dashboard_id = dashboard.id)",
			"DELETE FROM dashboard WHERE group_id = ?",
			"DELETE FROM group_user WHERE group_id = ?",
			"DELETE FROM group WHERE id = ?",
		}

		for _, sql := range deletes {
			_, err := sess.Exec(sql, cmd.Id)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
