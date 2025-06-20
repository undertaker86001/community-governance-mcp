package config

type CommunityGovernanceConfig struct {
	GitHubToken    string `json:"githubToken"`
	OpenAIKey      string `json:"openaiKey"`
	ImageAPIKey    string `json:"imageApiKey"`
	ImageAPIURL    string `json:"imageApiUrl"`
	KnowledgeDBURL string `json:"knowledgeDbUrl"`
	RepoOwner      string `json:"repoOwner"`
	RepoName       string `json:"repoName"`

	// 新增LLM意图识别配置
	IntentLLM IntentLLMConfig `json:"intentLlm"`
}

type IntentLLMConfig struct {
	ServiceName string `json:"serviceName"`
	Domain      string `json:"domain"`
	Port        int64  `json:"port"`
	Path        string `json:"path"`
	Model       string `json:"model"`
	APIKey      string `json:"apiKey"`
	Timeout     uint32 `json:"timeout"`
}
