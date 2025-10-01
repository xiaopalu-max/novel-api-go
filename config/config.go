package config

// ChatRequest 定义请求结构体
type ChatRequest struct {
	Authorization string    `json:"Authorization"`
	Messages      []Message `json:"messages"`
	Model         string    `json:"model"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Config struct {
	// 启动端口号变量
	Server struct {
		Addr string `yaml:"addr"`
	}

	// 日志管理密码
	LogsAdmin struct {
		Password string `yaml:"password"`
	} `yaml:"logs_admin"`

	// 存储桶选择器配置
	COS struct {
		Bucket string `yaml:"backet"` // 注意这里保持和.env文件中的拼写一致
	} `yaml:"cos"`

	// 翻译服务变量
	Translation struct {
		URL    string `yaml:"url"`
		Key    string `yaml:"key"`
		Model  string `yaml:"model"`
		Role   string `yaml:"role"`
		Enable bool   `yaml:"enable"`
	} `yaml:"translation"`

	// 腾讯云COS配置变量
	TencentCOS struct {
		SecretID  string `yaml:"secret_id"`
		SecretKey string `yaml:"secret_key"`
		Region    string `yaml:"region"`
		Bucket    string `yaml:"bucket"`
		BaseURL   string `yaml:"base_url"`
	} `yaml:"tencent_cos"`

	// Minio配置变量
	Minio struct {
		Endpoint        string `yaml:"endpoint"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
		BucketName      string `yaml:"bucket_name"`
		UseSSL          bool   `yaml:"use_ssl"`
		BaseURL         string `yaml:"base_url"`
	} `yaml:"minio"`

	// Alist配置变量
	Alist struct {
		BaseURL  string `yaml:"base_url"`
		Token    string `yaml:"token"`
		Path     string `yaml:"path"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"alist"`

	// 图片质量变量
	Parameters struct {
		ParamsVersion                      int     `yaml:"params_version"`
		Width                              int     `yaml:"width"`
		Height                             int     `yaml:"height"`
		Scale                              float64 `yaml:"scale"`
		Sampler                            string  `yaml:"sampler"`
		Steps                              int     `yaml:"steps"`
		NSamples                           int     `yaml:"n_samples"`
		UCPreset                           int     `yaml:"ucPreset"`
		QualityToggle                      bool    `yaml:"qualityToggle"`
		SM                                 bool    `yaml:"sm"`
		SMDyn                              bool    `yaml:"sm_dyn"`
		DynamicThresholding                bool    `yaml:"dynamic_thresholding"`
		ControlnetStrength                 int     `yaml:"controlnet_strength"`
		Legacy                             bool    `yaml:"legacy"`
		AddOriginalImage                   bool    `yaml:"add_original_image"`
		CFGRescale                         int     `yaml:"cfg_rescale"`
		NoiseSchedule                      string  `yaml:"noise_schedule"`
		LegacyV3Extend                     bool    `yaml:"legacy_v3_extend"`
		SkipCFGAboveSigma                  int     `yaml:"skip_cfg_above_sigma"`
		DeliberateEulerAncestralBug        bool    `yaml:"deliberate_euler_ancestral_bug"`
		PreferBrownian                     bool    `yaml:"prefer_brownian"`
		CustomAntiWords                    string  `yaml:"custom_anti_words"`
		AutoSmea                           bool    `yaml:"autoSmea"`
		UseCoords                          bool    `yaml:"use_coords"`
		LegacyUC                           bool    `yaml:"legacy_uc"`
		NormalizeReferenceStrengthMultiple bool    `yaml:"normalize_reference_strength_multiple"`
		InpaintImg2ImgStrength             int     `yaml:"inpaintImg2ImgStrength"`
		UseNewSharedTrial                  bool    `yaml:"use_new_shared_trial"`
	} `yaml:"parameters"`
}
