package sqlstore

import (
	"fmt"
	"time"

	"github.com/go-xorm/xorm"

	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
)

func init() {
	bus.AddHandler("sql", AddGroupUser)
	bus.AddHandler("sql", RemoveGroupUser)
	bus.AddHandler("sql", GetGroupUsers)
	bus.AddHandler("sql", UpdateGroupUser)
}

func AddGroupUser(cmd *m.AddGroupUserCommand) error {
	return inTransaction(func(sess *xorm.Session) error {
		// check if user exists
		if res, err := sess.Query("SELECT 1 from group_user WHERE group_id=? and user_id=?", cmd.GroupId, cmd.UserId); err != nil {
			return err
		} else if len(res) == 1 {
			return m.ErrGroupUserAlreadyAdded
		}

		entity := m.GroupUser{
			GroupId: cmd.GroupId,
			UserId:  cmd.UserId,
			Role:    cmd.Role,
			Created: time.Now(),
			Updated: time.Now(),
		}

		_, err := sess.Insert(&entity)
		return err
	})
}

func UpdateGroupUser(cmd *m.UpdateGroupUserCommand) error {
	return inTransaction(func(sess *xorm.Session) error {
		var groupUser m.GroupUser
		exists, err := sess.Where("group_id=? AND user_id=?", cmd.GroupId, cmd.UserId).Get(&groupUser)
		if err != nil {
			return err
		}

		if !exists {
			return m.ErrGroupUserNotFound
		}

		groupUser.Role = cmd.Role
		groupUser.Updated = time.Now()
		_, err = sess.Id(groupUser.Id).Update(&groupUser)
		if err != nil {
			return err
		}

		return validateOneAdminLeftInGroup(cmd.GroupId, sess)
	})
}

func GetGroupUsers(query *m.GetGroupUsersQuery) error {
	query.Result = make([]*m.GroupUserDTO, 0)
	sess := x.Table("group_user")
	sess.Join("INNER", "user", fmt.Sprintf("group_user.user_id=%s.id", x.Dialect().Quote("user")))
	sess.Where("group_user.group_id=?", query.GroupId)
	sess.Cols("group_user.group_id", "group_user.user_id", "user.email", "user.login", "group_user.role")
	sess.Asc("user.email", "user.login")

	err := sess.Find(&query.Result)
	return err
}

func RemoveGroupUser(cmd *m.RemoveGroupUserCommand) error {
	return inTransaction(func(sess *xorm.Session) error {
		var rawSql = "DELETE FROM group_user WHERE group_id=? and user_id=?"
		_, err := sess.Exec(rawSql, cmd.GroupId, cmd.UserId)
		if err != nil {
			return err
		}

		return validateOneAdminLeftInGroup(cmd.GroupId, sess)
	})
}

func validateOneAdminLeftInGroup(groupId int64, sess *xorm.Session) error {
	// validate that there is an admin user left
	res, err := sess.Query("SELECT 1 from group_user WHERE group_id=? and role='Admin'", groupId)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return m.ErrLastGroupAdmin
	}

	return err
}
