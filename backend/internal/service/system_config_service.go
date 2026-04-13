package service

import (
	"errors"
	"strings"

	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"
	"ai-blog/backend/internal/support"
)

type SystemConfigInput struct {
	Key   string
	Value string
}

type SystemConfigService struct {
	repo repository.SystemConfigRepository
}

func NewSystemConfigService(repo repository.SystemConfigRepository) *SystemConfigService {
	return &SystemConfigService{repo: repo}
}

func (service *SystemConfigService) ListAdmin() ([]map[string]any, error) {
	items, err := service.repo.ListAll()
	if err != nil {
		return nil, err
	}

	return service.mergeDefinitions(items), nil
}

func (service *SystemConfigService) Save(inputs []SystemConfigInput) ([]map[string]any, error) {
	definitionMap := support.DefaultSystemConfigMap()
	rows := make([]model.SystemConfig, 0, len(inputs))

	for _, item := range inputs {
		definition, exists := definitionMap[item.Key]
		if !exists {
			return nil, errors.New("存在不支持的系统配置项")
		}

		rows = append(rows, model.SystemConfig{
			Key:         definition.Key,
			Group:       definition.Group,
			Name:        definition.Name,
			Value:       strings.TrimSpace(item.Value),
			InputType:   definition.InputType,
			Description: definition.Description,
			IsPublic:    definition.IsPublic,
		})
	}

	if err := service.repo.SaveMany(rows); err != nil {
		return nil, err
	}

	return service.ListAdmin()
}

func (service *SystemConfigService) PublicMap() (map[string]string, error) {
	items, err := service.repo.ListPublic()
	if err != nil {
		return nil, err
	}

	definitions := support.DefaultSystemConfigDefinitions()
	valueMap := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		if definition.IsPublic {
			valueMap[definition.Key] = definition.Value
		}
	}

	for _, item := range items {
		if item.IsPublic {
			valueMap[item.Key] = item.Value
		}
	}

	return valueMap, nil
}

func (service *SystemConfigService) mergeDefinitions(items []model.SystemConfig) []map[string]any {
	existingMap := make(map[string]model.SystemConfig, len(items))
	for _, item := range items {
		existingMap[item.Key] = item
	}

	definitions := support.DefaultSystemConfigDefinitions()
	result := make([]map[string]any, 0, len(definitions))

	for _, definition := range definitions {
		value := definition.Value
		if item, exists := existingMap[definition.Key]; exists {
			value = item.Value
		}

		result = append(result, map[string]any{
			"key":           definition.Key,
			"group":         definition.Group,
			"name":          definition.Name,
			"value":         value,
			"default_value": definition.Value,
			"input_type":    definition.InputType,
			"description":   definition.Description,
			"is_public":     definition.IsPublic,
		})
	}

	return result
}
