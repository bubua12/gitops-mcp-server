package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
)

// ActionsService CI/CD 操作服务
type ActionsService struct {
	client *Client
}

// WorkflowInfo 工作流信息
type WorkflowInfo struct {
	ID        int64
	Name      string
	Path      string
	State     string
	URL       string
	CreatedAt string
	UpdatedAt string
}

// WorkflowRunInfo 工作流运行信息
type WorkflowRunInfo struct {
	ID           int64
	Name         string
	WorkflowID   int64
	Status       string
	Conclusion   string
	Branch       string
	CommitSHA    string
	CommitMsg    string
	URL          string
	HTMLURL      string
	CreatedAt    string
	UpdatedAtAt  string
	RunStartedAt string
	JobsURL      string
	LogsURL      string
}

// JobInfo Job 信息
type JobInfo struct {
	ID         int64
	Name       string
	Status     string
	Conclusion string
	StartedAt  string
	CompletedAt string
	Steps      []*StepInfo
}

// StepInfo Step 信息
type StepInfo struct {
	Name       string
	Status     string
	Conclusion string
	Number     int
}

// ListWorkflows 列出工作流
func (s *ActionsService) ListWorkflows(ctx context.Context, owner, repo string) ([]*WorkflowInfo, error) {
	owner = s.client.resolveOwner(owner)
	workflows, _, err := s.client.gh.Actions.ListWorkflows(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("list workflows %s/%s: %w", owner, repo, err)
	}
	var result []*WorkflowInfo
	for _, w := range workflows.Workflows {
		result = append(result, &WorkflowInfo{
			ID:        w.GetID(),
			Name:      w.GetName(),
			Path:      w.GetPath(),
			State:     w.GetState(),
			URL:       w.GetURL(),
			CreatedAt: w.GetCreatedAt().String(),
			UpdatedAt: w.GetUpdatedAt().String(),
		})
	}
	return result, nil
}

// ListWorkflowRuns 列出工作流运行记录
func (s *ActionsService) ListWorkflowRuns(ctx context.Context, owner, repo string, opts *github.ListWorkflowRunsOptions) ([]*WorkflowRunInfo, error) {
	owner = s.client.resolveOwner(owner)
	runs, _, err := s.client.gh.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, "", opts)
	if err != nil {
		return nil, fmt.Errorf("list workflow runs %s/%s: %w", owner, repo, err)
	}
	var result []*WorkflowRunInfo
	for _, r := range runs.WorkflowRuns {
		result = append(result, workflowRunToInfo(r))
	}
	return result, nil
}

// GetWorkflowRun 获取工作流运行详情
func (s *ActionsService) GetWorkflowRun(ctx context.Context, owner, repo string, runID int64) (*WorkflowRunInfo, error) {
	owner = s.client.resolveOwner(owner)
	run, _, err := s.client.gh.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return nil, fmt.Errorf("get workflow run %s/%s/%d: %w", owner, repo, runID, err)
	}
	return workflowRunToInfo(run), nil
}

// ListJobs 列出工作流运行的 Jobs
func (s *ActionsService) ListJobs(ctx context.Context, owner, repo string, runID int64) ([]*JobInfo, error) {
	owner = s.client.resolveOwner(owner)
	jobs, _, err := s.client.gh.Actions.ListWorkflowJobs(ctx, owner, repo, runID, nil)
	if err != nil {
		return nil, fmt.Errorf("list jobs for run %d: %w", runID, err)
	}
	var result []*JobInfo
	for _, j := range jobs.Jobs {
		job := &JobInfo{
			ID:          j.GetID(),
			Name:        j.GetName(),
			Status:      j.GetStatus(),
			Conclusion:  j.GetConclusion(),
			StartedAt:   j.GetStartedAt().String(),
			CompletedAt: j.GetCompletedAt().String(),
		}
		for _, step := range j.Steps {
			job.Steps = append(job.Steps, &StepInfo{
				Name:       step.GetName(),
				Status:     step.GetStatus(),
				Conclusion: step.GetConclusion(),
				Number:     int(step.GetNumber()),
			})
		}
		result = append(result, job)
	}
	return result, nil
}

// TriggerWorkflow 触发工作流
func (s *ActionsService) TriggerWorkflow(ctx context.Context, owner, repo string, workflowID int64, ref string, inputs map[string]any) error {
	owner = s.client.resolveOwner(owner)
	_, err := s.client.gh.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, "", github.CreateWorkflowDispatchEventRequest{
		Ref:    ref,
		Inputs: inputs,
	})
	if err != nil {
		return fmt.Errorf("trigger workflow %s/%s: %w", owner, repo, err)
	}
	return nil
}

// RerunWorkflow 重新运行工作流
func (s *ActionsService) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	owner = s.client.resolveOwner(owner)
	_, err := s.client.gh.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("rerun workflow %s/%s/%d: %w", owner, repo, runID, err)
	}
	return nil
}

// CancelWorkflow 取消工作流
func (s *ActionsService) CancelWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	owner = s.client.resolveOwner(owner)
	_, err := s.client.gh.Actions.CancelWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return fmt.Errorf("cancel workflow %s/%s/%d: %w", owner, repo, runID, err)
	}
	return nil
}

func workflowRunToInfo(r *github.WorkflowRun) *WorkflowRunInfo {
	return &WorkflowRunInfo{
		ID:           r.GetID(),
		Name:         r.GetName(),
		WorkflowID:   r.GetWorkflowID(),
		Status:       r.GetStatus(),
		Conclusion:   r.GetConclusion(),
		Branch:       r.GetHeadBranch(),
		CommitSHA:    r.GetHeadSHA(),
		CommitMsg:    r.GetHeadCommit().GetMessage(),
		URL:          r.GetURL(),
		HTMLURL:      r.GetHTMLURL(),
		CreatedAt:    r.GetCreatedAt().String(),
		RunStartedAt: r.GetRunStartedAt().String(),
		JobsURL:      r.GetJobsURL(),
		LogsURL:      r.GetLogsURL(),
	}
}

// FormatWorkflowList 格式化工作流列表
func FormatWorkflowList(workflows []*WorkflowInfo, owner, repo string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚙️ Workflows: %s/%s\n\n", owner, repo))
	if len(workflows) == 0 {
		sb.WriteString("无工作流\n")
		return sb.String()
	}
	for _, w := range workflows {
		state := "🟢"
		switch w.State {
		case "disabled_fork", "disabled_inactivity":
			state = "⚪"
		case "active":
			state = "🟢"
		default:
			state = "🟡"
		}
		sb.WriteString(fmt.Sprintf("%s **%s** (%s)\n  路径: %s\n", state, w.Name, w.State, w.Path))
	}
	return sb.String()
}

// FormatWorkflowRunList 格式化工作流运行列表
func FormatWorkflowRunList(runs []*WorkflowRunInfo, owner, repo string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🚀 Workflow Runs: %s/%s\n\n", owner, repo))
	if len(runs) == 0 {
		sb.WriteString("无运行记录\n")
		return sb.String()
	}
	for _, r := range runs {
		icon := getStatusIcon(r.Status, r.Conclusion)
		sb.WriteString(fmt.Sprintf("%s **%s** #%d [%s] %s\n  分支: %s | %s\n",
			icon, r.Name, r.ID, r.Status, r.Conclusion, r.Branch, r.HTMLURL))
	}
	return sb.String()
}

// FormatWorkflowRunDetail 格式化工作流运行详情
func FormatWorkflowRunDetail(run *WorkflowRunInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Workflow Run #%d\n\n", run.ID))
	sb.WriteString(fmt.Sprintf("- **名称:** %s\n", run.Name))
	sb.WriteString(fmt.Sprintf("- **状态:** %s (%s)\n", run.Status, run.Conclusion))
	sb.WriteString(fmt.Sprintf("- **分支:** %s\n", run.Branch))
	sb.WriteString(fmt.Sprintf("- **Commit:** %s\n", run.CommitSHA[:7]))
	if run.CommitMsg != "" {
		sb.WriteString(fmt.Sprintf("- **提交信息:** %s\n", run.CommitMsg))
	}
	sb.WriteString(fmt.Sprintf("- **链接:** %s\n", run.HTMLURL))
	sb.WriteString(fmt.Sprintf("- **创建时间:** %s\n", run.CreatedAt))
	return sb.String()
}

// FormatJobList 格式化 Job 列表
func FormatJobList(jobs []*JobInfo) string {
	var sb strings.Builder
	sb.WriteString("🔧 Jobs\n\n")
	if len(jobs) == 0 {
		sb.WriteString("无 Job 记录\n")
		return sb.String()
	}
	for _, j := range jobs {
		icon := getStatusIcon(j.Status, j.Conclusion)
		sb.WriteString(fmt.Sprintf("%s **%s** [%s/%s]\n", icon, j.Name, j.Status, j.Conclusion))
		if len(j.Steps) > 0 {
			for _, step := range j.Steps {
				stepIcon := getStatusIcon(step.Status, step.Conclusion)
				sb.WriteString(fmt.Sprintf("  %s %d. %s\n", stepIcon, step.Number, step.Name))
			}
		}
	}
	return sb.String()
}

func getStatusIcon(status, conclusion string) string {
	if status == "completed" {
		switch conclusion {
		case "success":
			return "✅"
		case "failure":
			return "❌"
		case "cancelled":
			return "⚪"
		case "skipped":
			return "⏭️"
		default:
			return "🟡"
		}
	}
	if status == "in_progress" {
		return "🔄"
	}
	return "⏳"
}
