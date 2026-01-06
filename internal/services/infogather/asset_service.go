package infogather

import (
	"JAttack/internal/db"
	"context"
)

type AssetService struct {
	ctx       context.Context
	dbManager *db.Manager
}

func NewAssetService(dbManager *db.Manager) *AssetService {
	return &AssetService{
		dbManager: dbManager,
	}
}

func (s *AssetService) Startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *AssetService) GetAssets() ([]db.Asset, error) {
	return s.dbManager.GetAllAssets()
}

func (s *AssetService) GetAsset(id int64) (*db.Asset, error) {
	return s.dbManager.GetAsset(id)
}

func (s *AssetService) GetAssetPorts(assetID int64) ([]db.AssetPort, error) {
	return s.dbManager.GetAssetPorts(assetID)
}

func (s *AssetService) GetWebServices(assetID int64) ([]db.WebService, error) {
	return s.dbManager.GetWebServices(assetID)
}

func (s *AssetService) GetWebDirectories(webServiceID int64) ([]db.WebDirectory, error) {
	return s.dbManager.GetWebDirectories(webServiceID)
}

func (s *AssetService) GetAssetWebDirectories(assetID int64) ([]db.WebDirectory, error) {
	return s.dbManager.GetAssetWebDirectories(assetID)
}

func (s *AssetService) GetWebJSFiles(webServiceID int64) ([]db.WebJSFile, error) {
	return s.dbManager.GetWebJSFiles(webServiceID)
}

func (s *AssetService) GetAssetWebJSFiles(assetID int64) ([]db.WebJSFile, error) {
	return s.dbManager.GetAssetWebJSFiles(assetID)
}

func (s *AssetService) GetAssetSensitiveResults(assetID int64) ([]db.SensitiveResult, error) {
	return s.dbManager.GetAssetSensitiveResults(assetID)
}

func (s *AssetService) GetAssetAuthResults(assetID int64) ([]db.AuthResult, error) {
	return s.dbManager.GetAssetAuthResults(assetID)
}
