package agent

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"personal-doctor/backend/internal/config"
	"personal-doctor/backend/internal/domain"
)

type DoctorAgent struct {
	provider string
	model    *openai.ChatModel
}

func NewDoctorAgent(cfg config.AgentConfig) (*DoctorAgent, error) {
	if cfg.Provider != "openai" || cfg.OpenAIAPIKey == "" {
		return &DoctorAgent{provider: "mock"}, nil
	}

	modelCfg := &openai.ChatModelConfig{
		APIKey: cfg.OpenAIAPIKey,
		Model:  cfg.OpenAIModel,
	}
	if cfg.OpenAIBaseURL != "" {
		modelCfg.BaseURL = normalizeOpenAIBaseURL(cfg.OpenAIBaseURL)
	}

	model, err := openai.NewChatModel(context.Background(), modelCfg)
	if err != nil {
		return nil, err
	}
	return &DoctorAgent{provider: "openai", model: model}, nil
}

func (a *DoctorAgent) Chat(ctx context.Context, input domain.ChatInput) (string, error) {
	if a.provider == "mock" {
		return mockReply(input), nil
	}

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt(input)),
	}
	for _, item := range input.History {
		switch item.Role {
		case "user":
			messages = append(messages, schema.UserMessage(item.Content))
		case "assistant":
			messages = append(messages, schema.AssistantMessage(item.Content, nil))
		}
	}
	messages = append(messages, schema.UserMessage(input.Message))

	response, err := a.model.Generate(ctx, messages)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response.Content), nil
}

func systemPrompt(input domain.ChatInput) string {
	var builder strings.Builder
	builder.WriteString(`你是一个谨慎、专业的个人医生聊天助手，用中文回答。
你可以帮助用户整理病情、解释医学概念、提示复诊/检查/用药注意事项。
安全边界：
- 不直接替代医生诊断，不擅自开处方或调整处方剂量。
- 遇到胸痛、呼吸困难、意识障碍、大出血、严重过敏、疑似中风等急症，明确建议立即就医或拨打急救电话。
- 对药物相互作用、孕产、儿童、慢病、肝肾功能异常等高风险场景要提示咨询医生/药师。
- 回答应结合已录入病历，并说明哪些信息仍缺失。`)
	builder.WriteString("\n\n当前病人：\n")
	builder.WriteString(fmt.Sprintf("姓名：%s\n性别：%s\n生日：%s\n电话：%s\n过敏史：%s\n备注：%s\n",
		empty(input.Patient.Name), empty(input.Patient.Gender), empty(input.Patient.Birthday), empty(input.Patient.Phone), empty(input.Patient.Allergies), empty(input.Patient.Notes)))

	builder.WriteString("\n已录入病历/药方：\n")
	if len(input.Records) == 0 {
		builder.WriteString("暂无。\n")
		return builder.String()
	}
	for i, record := range input.Records {
		if i >= 20 {
			builder.WriteString("其余较早记录已省略。\n")
			break
		}
		builder.WriteString(fmt.Sprintf("- [%s] %s：%s\n%s\n", record.Kind, record.Title, record.RecordedAt.Format("2006-01-02"), record.Content))
	}
	return builder.String()
}

func empty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未填写"
	}
	return value
}

func normalizeOpenAIBaseURL(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Path == "" || parsed.Path == "/" {
		return trimmed + "/v1"
	}
	if strings.HasSuffix(parsed.Path, "/v1") {
		return trimmed
	}
	return trimmed + "/v1"
}

func mockReply(input domain.ChatInput) string {
	recordCount := len(input.Records)
	return fmt.Sprintf(`我已收到你关于「%s」的描述，并会结合 %s 的 %d 条病历/药方记录来整理。

目前后端还没有配置真实模型 API，所以这是本地占位回复。你后面给我 API 后，我会把请求交给 Eino 的 ChatModel。

建议你继续补充：
1. 症状开始时间、持续多久、是否加重
2. 体温、血压、心率等可测指标
3. 正在使用的药物名称、剂量、频次
4. 既往病史、过敏史、近期检查结果

如果出现胸痛、呼吸困难、意识模糊、严重过敏、大出血等急症，请立即线下就医或拨打急救电话。`, input.Message, input.Patient.Name, recordCount)
}
