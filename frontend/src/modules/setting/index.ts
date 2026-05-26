import { apiClient } from '../../lib/axios';
import type { Config } from './type';
import { API_ENDPOINTS } from '../../constant/endpoint';

export async function getConfig(): Promise<Config> {
  try {
    const response = await apiClient.get(API_ENDPOINTS.SETTINGS.BASE);
    const config = response.data.data;
    
    // Konversi snake_case dari backend ke camelCase untuk komponen frontend
    return {
      erpType: config.erp_type,
      endpoint: config.endpoint,
      apiKey: config.api_key,
      volumeSensitivity: config.volume_sensitivity,
      geoThreshold: config.geo_threshold,
      velocityLimit: config.velocity_limit,
    };
  } catch (error) {
    console.error('Gagal memuat pengaturan:', error);
    throw error;
  }
}

export async function saveConfig(config: Config): Promise<Config> {
  try {
    // Konversi camelCase ke snake_case yang diharapkan backend
    const payload = {
      erp_type: config.erpType,
      endpoint: config.endpoint,
      api_key: config.apiKey,
      volume_sensitivity: config.volumeSensitivity,
      geo_threshold: config.geoThreshold,
      velocity_limit: config.velocityLimit,
    };
    
    const response = await apiClient.put(API_ENDPOINTS.SETTINGS.BASE, payload);
    const updatedConfig = response.data.data;

    return {
      erpType: updatedConfig.erp_type,
      endpoint: updatedConfig.endpoint,
      apiKey: updatedConfig.api_key,
      volumeSensitivity: updatedConfig.volume_sensitivity,
      geoThreshold: updatedConfig.geo_threshold,
      velocityLimit: updatedConfig.velocity_limit,
    };
  } catch (error) {
    console.error('Gagal menyimpan pengaturan:', error);
    throw error;
  }
}
