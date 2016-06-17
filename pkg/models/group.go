package models

import (
        "errors"
        "time"
)

// Typed errors
var (
        ErrGroupNotFound  = errors.New("Group not found")
        ErrGroupNameTaken = errors.New("Group name is taken")
)

type Group struct {
        Id      int64
        Name    string

	Created time.Time
        Updated time.Time
}

// ---------------------
// COMMANDS

type CreateGroupCommand struct {
        Name string `json:"name" binding:"Required"`

        // initial admin user for account
        UserId int64 `json:"-"`
        Result Group   `json:"-"`
}

type DeleteGroupCommand struct {
        Id int64
}

type UpdateGroupCommand struct {
        Name  string
        GroupId int64
}

type GetGroupByIdQuery struct {
        Id     int64
        Result *Group
}

type GetGroupByNameQuery struct {
        Name   string
        Result *Group
}

type SearchGroupsQuery struct {
        Query string
        Name  string
        Limit int
        Page  int

        Result []*GroupDTO
}

type GroupDTO struct {
        Id   int64  `json:"id"`
        Name string `json:"name"`
}

type UserGroupDTO struct {
        GroupId int64    `json:"groupId"`
        Name  string   `json:"name"`
        Role  RoleType `json:"role"`
}
