package google

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// GroupsClient Google Groups客户端
type GroupsClient struct {
	service *admin.Service
	config  *GroupsConfig
}

// NewGroupsClient 创建Groups客户端
func NewGroupsClient(config *GroupsConfig) (*GroupsClient, error) {
	// 读取服务账号凭证
	credentials, err := ioutil.ReadFile("credentials.json") // 使用默认凭证文件
	if err != nil {
		return nil, fmt.Errorf("无法读取凭证文件: %v", err)
	}

	// 创建JWT配置
	jwtConfig, err := google.JWTConfigFromJSON(credentials, admin.AdminDirectoryGroupReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("无法创建JWT配置: %v", err)
	}

	// 设置委托用户
	jwtConfig.Subject = config.AdminEmail

	// 创建HTTP客户端
	client := jwtConfig.Client(context.Background())

	// 创建Admin Directory服务
	service, err := admin.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("无法创建Admin Directory服务: %v", err)
	}

	return &GroupsClient{
		service: service,
		config:  config,
	}, nil
}

// GetGroupMembers 获取邮件组成员
func (c *GroupsClient) GetGroupMembers(groupKey string) ([]string, error) {
	var members []string
	pageToken := ""

	for {
		// 获取组成员列表
		response, err := c.service.Members.List(groupKey).PageToken(pageToken).Do()
		if err != nil {
			return nil, fmt.Errorf("获取组成员失败: %v", err)
		}

		// 提取成员邮箱
		for _, member := range response.Members {
			if member.Email != "" {
				members = append(members, member.Email)
			}
		}

		// 检查是否有下一页
		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	return members, nil
}

// AddGroupMember 添加邮件组成员
func (c *GroupsClient) AddGroupMember(groupKey, email string) error {
	member := &admin.Member{
		Email: email,
		Role:  "MEMBER",
	}

	_, err := c.service.Members.Insert(groupKey, member).Do()
	if err != nil {
		return fmt.Errorf("添加组成员失败: %v", err)
	}

	return nil
}

// RemoveGroupMember 移除邮件组成员
func (c *GroupsClient) RemoveGroupMember(groupKey, email string) error {
	err := c.service.Members.Delete(groupKey, email).Do()
	if err != nil {
		return fmt.Errorf("移除组成员失败: %v", err)
	}

	return nil
}

// GetGroupInfo 获取邮件组信息
func (c *GroupsClient) GetGroupInfo(groupKey string) (*admin.Group, error) {
	group, err := c.service.Groups.Get(groupKey).Do()
	if err != nil {
		return nil, fmt.Errorf("获取邮件组信息失败: %v", err)
	}

	return group, nil
}

// ListGroups 列出所有邮件组
func (c *GroupsClient) ListGroups(domain string) ([]*admin.Group, error) {
	var groups []*admin.Group
	pageToken := ""

	for {
		// 获取邮件组列表
		response, err := c.service.Groups.List().Domain(domain).PageToken(pageToken).Do()
		if err != nil {
			return nil, fmt.Errorf("获取邮件组列表失败: %v", err)
		}

		groups = append(groups, response.Groups...)

		// 检查是否有下一页
		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	return groups, nil
}

// CreateGroup 创建邮件组
func (c *GroupsClient) CreateGroup(name, email, description string) (*admin.Group, error) {
	group := &admin.Group{
		Email:       email,
		Name:        name,
		Description: description,
	}

	createdGroup, err := c.service.Groups.Insert(group).Do()
	if err != nil {
		return nil, fmt.Errorf("创建邮件组失败: %v", err)
	}

	return createdGroup, nil
}

// UpdateGroupSettings 更新邮件组设置
func (c *GroupsClient) UpdateGroupSettings(groupKey string, settings map[string]interface{}) error {
	// 这里可以添加更新邮件组设置的逻辑
	// 具体实现取决于需要更新的设置类型
	log.Printf("更新邮件组设置: %s", groupKey)
	return nil
}

// GetGroupSettings 获取邮件组设置
func (c *GroupsClient) GetGroupSettings(groupKey string) (map[string]interface{}, error) {
	// 这里可以添加获取邮件组设置的逻辑
	// 具体实现取决于需要获取的设置类型
	log.Printf("获取邮件组设置: %s", groupKey)
	return make(map[string]interface{}), nil
}
