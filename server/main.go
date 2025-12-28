package main

import (
	"context"
	"fmt"
	"log"

	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

type WriteInput struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

type ReadInput struct {
	FileName string `json:"file_name"`
}

type ListDirectoryInp struct {
	DirPath string `json:"dir_path"`
}

func write_file(ctx tool.Context, input WriteInput) (string, error) {
	file, err := os.OpenFile(input.FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "Failed to open/create file " + input.FileName, err
	}
	defer file.Close()
	if _, err := file.WriteString(input.Content); err != nil {
		return "Failed to write the content to file " + input.FileName, err
	}
	return "Successfully written the content to file " + input.Content, nil
}

func read_file(ctx tool.Context, input ReadInput) (string, error) {
	if content_bytes, err := os.ReadFile(input.FileName); err != nil {
		return "Failed to read the content of the file " + input.FileName, err
	} else {
		return string(content_bytes), nil
	}
}

func list_directory(ctx tool.Context, input ListDirectoryInp) (string, error) {
	if files, err := os.ReadDir(input.DirPath); err != nil {
		return "Error listing dir: " + input.DirPath, err
	} else {
		s := "FileName\tIsDir\n\n"
		for _, file := range files {
			s += fmt.Sprintf("%s\t%s", file.Name(), file.IsDir())
		}
		return s, nil
	}
}

func main() {
	ctx := context.Background()

	model, err := gemini.NewModel(ctx, "gemini-2.5-pro", &genai.ClientConfig{APIKey: os.Getenv("GOOGLE_API_KEY")})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	write_file_tool, err := functiontool.New(functiontool.Config{
		Name:        "write_file",
		Description: "Writes content to the requested file.",
	},
		write_file,
	)

	if err != nil {
		log.Fatalf("Failed to create write_file tool %v", err)
	}

	read_file_tool, err := functiontool.New(functiontool.Config{
		Name:        "read_file",
		Description: "Reads content of the requested file.",
	},
		read_file,
	)

	if err != nil {
		log.Fatalf("Failed to create read_file tool %v", err)
	}

	list_directory_tool, err := functiontool.New(functiontool.Config{
		Name:        "list_directory",
		Description: "List the directory and returs file name and if it is a dir",
	},
		list_directory,
	)

	if err != nil {
		log.Fatalf("Failed to create read_file tool %v", err)
	}

	analyzerAgent, err := llmagent.New(llmagent.Config{
		Name:        "analyzer_agent",
		Model:       model,
		Description: "Analyzes user request and the code base",
		Instruction: `You are excellent at understanding the user's request. Understand what user wants, analyze the code base by reading only the necessary files and genearting analytical thoughts about each file. 
			
		### Allowed Tools:	
			Use 'list_directory' tool to list the directory.
			It is always a good idea to explore the directory to get the idea of the project and files.

			Use 'read_file' tool to read the contents of the file.
			Always read the content of the files before analyzing and generating thoughts about it.

		Your job is to make the analysis with the given tools, create report for the planner agent to plan.

		Example output:
		## Analysis
		### User request: // whatever the user request is

		### file_name_1
		// your analysis of the file_name_1


		### file_name_2
		// your analysis of the file_name_2
		.
		.
		.

		### Final Thoughts
		// Your final thoughts
		.
		.
		.

		## End of analysis
		`,
		OutputKey: "analysis",
		Tools: []tool.Tool{
			read_file_tool,
			list_directory_tool,
		},
	})

	if err != nil {
		log.Fatalf("Failed to create analyzer_agent %v", err)
	}

	plannerAgent, err := llmagent.New(llmagent.Config{
		Name:        "planner_agent",
		Model:       model,
		Description: "Generates execution plan",
		Instruction: `You are excellent at planning software engineering tasks.
		Given is the analysis of what user wants and the repo analysis.

		### User Request and Repo Analysis
		{analysis}
		
		Generate a implementation plan in markdown bullet-points to resolve the user request.
		`,
		OutputKey: "implementation_plan",
	})

	if err != nil {
		log.Fatalf("Failed to create planner_agent %v", err)
	}

	codeAgent, err := llmagent.New(llmagent.Config{
		Name:        "code_agent",
		Model:       model,
		Description: "Writes code",
		Instruction: `You are an excellent programmer who helps user by writing code. 

		### Allowed Tools:	
			Use 'write_file' tool to write contents to the file. 
			Always write code to a file with the code without any extra lines, prefix or suffix.
			
			Use 'read_file' tool to read the contents of the file.
			Always read the content of existing file before writing to the the file.

			Use 'list_directory' tool to list the directory.
			It is always a good idea to explore the directory to get the idea of the project and files.


		Given is the user request, repo analysis and implementation plan, execute things accordingly.

		### Analysis

		{analysis}

		### Implementation Plan
		{implementation_plan}
		`,
		Tools: []tool.Tool{
			write_file_tool,
			read_file_tool,
			list_directory_tool,
		},
	})

	if err != nil {
		log.Fatalf("Failed to create code_agent %v", err)
	}

	codepipeline, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name: "CodePipelineAgent",
			SubAgents: []agent.Agent{
				analyzerAgent,
				plannerAgent,
				codeAgent,
			},
			Description: "Executes a software engineering task request.",
		},
	})

	rootagent := codepipeline

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(rootagent),
	}

	l := full.NewLauncher()
	if err := l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Failed to execute agent: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
