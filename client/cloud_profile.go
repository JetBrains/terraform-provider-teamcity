package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"terraform-provider-teamcity/models"
)

const (
	cloudProfileFeatureType       = "CloudProfile"
	cloudImageFeatureType         = "CloudImage"
	cloudProfileDiscoveryAttempts = 2
)

func (c *Client) CreateCloudProfile(projectID string, profile models.CloudProfileJson) (*models.CloudProfileJson, error) {
	before, err := c.getProjectFeatures(projectID)
	if err != nil {
		return nil, err
	}

	if err := c.createProjectFeature(projectID, cloudProfileFeature(profile)); err != nil {
		return nil, err
	}

	createdProfile, err := c.discoverCreatedCloudProfile(projectID, before)
	if err != nil {
		return nil, err
	}

	profileID := *createdProfile.Id
	for _, image := range images(profile) {
		if err := c.createProjectFeature(projectID, cloudImageFeature(image, profileID)); err != nil {
			return nil, c.cleanupFailedCloudProfileCreate(projectID, profileID, err)
		}
	}

	created, err := c.GetCloudProfile(projectID, profileID)
	if err != nil {
		return nil, c.cleanupFailedCloudProfileCreate(projectID, profileID, err)
	}
	if created == nil {
		return nil, c.cleanupFailedCloudProfileCreate(projectID, profileID, fmt.Errorf("created cloud profile project feature %q is unavailable", profileID))
	}
	return created, nil
}

func (c *Client) discoverCreatedCloudProfile(projectID string, before *models.ProjectFeaturesJson) (*models.ProjectFeatureJson, error) {
	var lastErr error
	for attempt := 1; attempt <= cloudProfileDiscoveryAttempts; attempt++ {
		after, err := c.getProjectFeatures(projectID)
		if err != nil {
			lastErr = err
			continue
		}
		createdProfile, ok := newlyCreatedFeature(before, after, cloudProfileFeatureType)
		if ok && createdProfile.Id != nil {
			return &createdProfile, nil
		}
		lastErr = fmt.Errorf("TeamCity did not return the created cloud profile project feature")
	}

	return nil, fmt.Errorf("TeamCity may have created an unmanaged cloud profile in project %q because its server-assigned ID could not be discovered after %d attempts: %w", projectID, cloudProfileDiscoveryAttempts, lastErr)
}

func (c *Client) cleanupFailedCloudProfileCreate(projectID, profileID string, createErr error) error {
	if err := c.DeleteCloudProfile(projectID, profileID); err != nil {
		return fmt.Errorf("%w; additionally failed to clean up cloud profile project feature %q: %v", createErr, profileID, err)
	}
	return createErr
}

func (c *Client) GetCloudProfile(projectID, profileID string) (*models.CloudProfileJson, error) {
	feature, err := c.getProjectFeature(projectID, profileID)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if feature.Type != cloudProfileFeatureType {
		return nil, fmt.Errorf("project feature %q is %q, not a cloud profile", profileID, feature.Type)
	}

	features, err := c.getProjectFeatures(projectID)
	if err != nil {
		return nil, err
	}
	profile := cloudProfileFromFeature(*feature)
	projectIDCopy := projectID
	profile.Project = &models.CloudProfileProjectJson{Id: &projectIDCopy}
	profile.Images = &models.CloudImagesJson{}
	for _, candidate := range features.ProjectFeature {
		if candidate.Type != cloudImageFeatureType || propertyValue(candidate.Properties, "profileId") != profileID {
			continue
		}
		profile.Images.CloudImage = append(profile.Images.CloudImage, cloudImageFromFeature(candidate))
	}
	return &profile, nil
}

func (c *Client) UpdateCloudProfile(projectID, profileID string, profile models.CloudProfileJson) (*models.CloudProfileJson, error) {
	if err := c.updateProjectFeature(projectID, profileID, cloudProfileFeature(profile)); err != nil {
		return nil, err
	}

	features, err := c.getProjectFeatures(projectID)
	if err != nil {
		return nil, err
	}
	existingImages := map[string]struct{}{}
	for _, feature := range features.ProjectFeature {
		if feature.Type == cloudImageFeatureType && feature.Id != nil && propertyValue(feature.Properties, "profileId") == profileID {
			existingImages[*feature.Id] = struct{}{}
		}
	}

	for _, image := range images(profile) {
		feature := cloudImageFeature(image, profileID)
		if image.Id == "" {
			if err := c.createProjectFeature(projectID, feature); err != nil {
				return nil, err
			}
			continue
		}
		if _, exists := existingImages[image.Id]; !exists {
			return nil, fmt.Errorf("cloud image project feature %q no longer exists", image.Id)
		}
		if err := c.updateProjectFeature(projectID, image.Id, feature); err != nil {
			return nil, err
		}
		delete(existingImages, image.Id)
	}
	for imageID := range existingImages {
		if err := c.DeleteProjectFeature(projectID, imageID); err != nil {
			return nil, err
		}
	}

	return c.GetCloudProfile(projectID, profileID)
}

func (c *Client) DeleteCloudProfile(projectID, profileID string) error {
	features, err := c.getProjectFeatures(projectID)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, feature := range features.ProjectFeature {
		if feature.Type == cloudImageFeatureType && feature.Id != nil && propertyValue(feature.Properties, "profileId") == profileID {
			if err := c.DeleteProjectFeature(projectID, *feature.Id); err != nil {
				return err
			}
		}
	}
	return c.DeleteProjectFeature(projectID, profileID)
}

// getProjectFeatures uses an explicit projection because generic project-feature reads
// can omit properties required to reconstruct Terraform state. Feature POST responses can
// also be empty, so createProjectFeature intentionally does not rely on a response body.
func (c *Client) getProjectFeatures(projectID string) (*models.ProjectFeaturesJson, error) {
	var features models.ProjectFeaturesJson
	if err := c.GetRequest(projectFeaturesEndpoint(projectID), "fields=projectFeature(id,type,properties(property(name,value)))", &features); err != nil {
		return nil, err
	}
	return &features, nil
}

func (c *Client) getProjectFeature(projectID, featureID string) (*models.ProjectFeatureJson, error) {
	var feature models.ProjectFeatureJson
	if err := c.GetRequest(projectFeatureEndpoint(projectID, featureID), "fields=id,type,properties(property(name,value))", &feature); err != nil {
		return nil, err
	}
	return &feature, nil
}

func (c *Client) createProjectFeature(projectID string, feature models.ProjectFeatureJson) error {
	body, err := json.Marshal(feature)
	if err != nil {
		return err
	}
	return c.PostRequest(projectFeaturesEndpoint(projectID), bytes.NewReader(body), nil)
}

func (c *Client) updateProjectFeature(projectID, featureID string, feature models.ProjectFeatureJson) error {
	body, err := json.Marshal(feature)
	if err != nil {
		return err
	}
	return c.PutRequest(projectFeatureEndpoint(projectID, featureID), bytes.NewReader(body), nil)
}

func projectFeaturesEndpoint(projectID string) string {
	return fmt.Sprintf("/projects/id:%s/projectFeatures", projectID)
}

func projectFeatureEndpoint(projectID, featureID string) string {
	return fmt.Sprintf("/projects/id:%s/projectFeatures/id:%s", projectID, featureID)
}

func cloudProfileFeature(profile models.CloudProfileJson) models.ProjectFeatureJson {
	properties := copyProperties(profile.Properties)
	setProperty(&properties, "cloud-code", profile.CloudProviderId)
	setProperty(&properties, "name", profile.Name)
	return models.ProjectFeatureJson{Type: cloudProfileFeatureType, Properties: properties}
}

func cloudImageFeature(image models.CloudImageJson, profileID string) models.ProjectFeatureJson {
	properties := copyProperties(image.Properties)
	setProperty(&properties, "profileId", profileID)
	setProperty(&properties, "image-name-prefix", image.Name)
	if image.AgentPoolId != nil {
		setProperty(&properties, "agent_pool_id", strconv.Itoa(*image.AgentPoolId))
	}
	return models.ProjectFeatureJson{Type: cloudImageFeatureType, Properties: properties}
}

func cloudProfileFromFeature(feature models.ProjectFeatureJson) models.CloudProfileJson {
	profile := models.CloudProfileJson{Properties: &models.Properties{Property: copyProperties(&feature.Properties).Property}}
	if feature.Id != nil {
		profile.Id = *feature.Id
	}
	profile.CloudProviderId = propertyValue(feature.Properties, "cloud-code")
	profile.Name = propertyValue(feature.Properties, "name")
	removeProperties(profile.Properties, "cloud-code", "name")
	return profile
}

func cloudImageFromFeature(feature models.ProjectFeatureJson) models.CloudImageJson {
	image := models.CloudImageJson{Properties: &models.Properties{Property: copyProperties(&feature.Properties).Property}}
	if feature.Id != nil {
		image.Id = *feature.Id
	}
	image.Name = propertyValue(feature.Properties, "image-name-prefix")
	if value := propertyValue(feature.Properties, "agent_pool_id"); value != "" {
		if poolID, err := strconv.Atoi(value); err == nil {
			image.AgentPoolId = &poolID
		}
	}
	removeProperties(image.Properties, "profileId", "image-name-prefix", "agent_pool_id")
	return image
}

func images(profile models.CloudProfileJson) []models.CloudImageJson {
	if profile.Images == nil {
		return nil
	}
	return profile.Images.CloudImage
}

func newlyCreatedFeature(before, after *models.ProjectFeaturesJson, featureType string) (models.ProjectFeatureJson, bool) {
	existing := map[string]struct{}{}
	for _, feature := range before.ProjectFeature {
		if feature.Id != nil {
			existing[*feature.Id] = struct{}{}
		}
	}
	for _, feature := range after.ProjectFeature {
		if feature.Type == featureType && feature.Id != nil {
			if _, found := existing[*feature.Id]; !found {
				return feature, true
			}
		}
	}
	return models.ProjectFeatureJson{}, false
}

func copyProperties(properties *models.Properties) models.Properties {
	if properties == nil {
		return models.Properties{}
	}
	result := models.Properties{Property: append([]models.Property(nil), properties.Property...)}
	sort.Slice(result.Property, func(i, j int) bool { return result.Property[i].Name < result.Property[j].Name })
	return result
}

func propertyValue(properties models.Properties, name string) string {
	for _, property := range properties.Property {
		if property.Name == name {
			return property.Value
		}
	}
	return ""
}

func setProperty(properties *models.Properties, name, value string) {
	for index := range properties.Property {
		if properties.Property[index].Name == name {
			properties.Property[index].Value = value
			return
		}
	}
	properties.Property = append(properties.Property, models.Property{Name: name, Value: value})
	sort.Slice(properties.Property, func(i, j int) bool { return properties.Property[i].Name < properties.Property[j].Name })
}

func removeProperties(properties *models.Properties, names ...string) {
	removed := map[string]struct{}{}
	for _, name := range names {
		removed[name] = struct{}{}
	}
	kept := properties.Property[:0]
	for _, property := range properties.Property {
		if _, exists := removed[property.Name]; !exists {
			kept = append(kept, property)
		}
	}
	properties.Property = kept
}
