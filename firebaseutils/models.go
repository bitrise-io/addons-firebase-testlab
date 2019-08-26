package firebaseutils

import (
	"github.com/bitrise-io/addons-firebase-testlab/configs"
)

// APIModel ...
type APIModel struct {
	JWT *configs.JWTModel
}

/*
// TestDeviceCatalog ...
type TestDeviceCatalog struct {
	AndroidDeviceCatalog struct {
		Models []struct {
			Brand               string   `json:"brand"`
			Codename            string   `json:"codename"`
			Form                string   `json:"form"`
			ID                  string   `json:"id"`
			Manufacturer        string   `json:"manufacturer"`
			Name                string   `json:"name"`
			ScreenDensity       int      `json:"screenDensity"`
			ScreenX             int      `json:"screenX"`
			ScreenY             int      `json:"screenY"`
			SupportedAbis       []string `json:"supportedAbis"`
			SupportedVersionIds []string `json:"supportedVersionIds,omitempty"`
			Tags                []string `json:"tags,omitempty"`
		} `json:"models"`
		RuntimeConfiguration struct {
			Locales []struct {
				ID     string   `json:"id"`
				Name   string   `json:"name"`
				Region string   `json:"region,omitempty"`
				Tags   []string `json:"tags,omitempty"`
			} `json:"locales"`
			Orientations []struct {
				ID   string   `json:"id"`
				Name string   `json:"name"`
				Tags []string `json:"tags,omitempty"`
			} `json:"orientations"`
		} `json:"runtimeConfiguration"`
		Versions []struct {
			APILevel    int    `json:"apiLevel"`
			CodeName    string `json:"codeName"`
			ID          string `json:"id"`
			ReleaseDate struct {
				Day   int `json:"day"`
				Month int `json:"month"`
				Year  int `json:"year"`
			} `json:"releaseDate"`
			VersionString string   `json:"versionString"`
			Tags          []string `json:"tags,omitempty"`
		} `json:"versions"`
	} `json:"androidDeviceCatalog"`
}

// StepsModel ...
type StepsModel struct {
	NextPageToken string      `json:"nextPageToken"`
	Steps         []StepModel `json:"steps"`
}

// StepModel ...
type StepModel struct {
	CompletionTime struct {
		Nanos   int `json:"nanos"`
		Seconds int `json:"seconds,string"`
	} `json:"completionTime"`
	CreationTime struct {
		Nanos   int `json:"nanos"`
		Seconds int `json:"seconds,string"`
	} `json:"creationTime"`
	Description         string `json:"description"`
	DeviceUsageDuration struct {
		Nanos   int `json:"nanos"`
		Seconds int `json:"seconds,string"`
	} `json:"deviceUsageDuration"`
	DimensionValue []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"dimensionValue"`
	HasImages bool `json:"hasImages"`
	Labels    []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"labels"`
	Name    string `json:"name"`
	Outcome struct {
		FailureDetail struct {
			Crashed          bool `json:"crashed"`
			NotInstalled     bool `json:"notInstalled"`
			OtherNativeCrash bool `json:"otherNativeCrash"`
			TimedOut         bool `json:"timedOut"`
			UnableToCrawl    bool `json:"unableToCrawl"`
		} `json:"failureDetail"`
		InconclusiveDetail struct {
			AbortedByUser         bool `json:"abortedByUser"`
			InfrastructureFailure bool `json:"infrastructureFailure"`
		} `json:"inconclusiveDetail"`
		SkippedDetail struct {
			IncompatibleAppVersion   bool `json:"incompatibleAppVersion"`
			IncompatibleArchitecture bool `json:"incompatibleArchitecture"`
			IncompatibleDevice       bool `json:"incompatibleDevice"`
		} `json:"skippedDetail"`
		SuccessDetail struct {
			OtherNativeCrash bool `json:"otherNativeCrash"`
		} `json:"successDetail"`
		Summary string `json:"summary"`
	} `json:"outcome"`
	RunDuration struct {
		Nanos   int `json:"nanos"`
		Seconds int `json:"seconds,string"`
	} `json:"runDuration"`
	State             string `json:"state"`
	StepID            string `json:"stepId"`
	TestExecutionStep struct {
		TestIssues []struct {
			ErrorMessage string `json:"errorMessage"`
			StackTrace   struct {
				Exception string `json:"exception"`
			} `json:"stackTrace"`
		} `json:"testIssues"`
		TestSuiteOverviews []struct {
			ErrorCount   int    `json:"errorCount"`
			FailureCount int    `json:"failureCount"`
			Name         string `json:"name"`
			SkippedCount int    `json:"skippedCount"`
			TotalCount   int    `json:"totalCount"`
			XMLSource    struct {
				FileURI string `json:"fileUri"`
			} `json:"xmlSource"`
		} `json:"testSuiteOverviews"`
		TestTiming struct {
			TestProcessDuration struct {
				Nanos   int `json:"nanos"`
				Seconds int `json:"seconds,string"`
			} `json:"testProcessDuration"`
		} `json:"testTiming"`
		ToolExecution struct {
			CommandLineArguments []string `json:"commandLineArguments"`
			ExitCode             struct {
				Number int `json:"number"`
			} `json:"exitCode"`
			ToolLogs []struct {
				FileURI string `json:"fileUri"`
			} `json:"toolLogs"`
			ToolOutputs []struct {
				CreationTime struct {
					Nanos   int `json:"nanos"`
					Seconds int `json:"seconds,string"`
				} `json:"creationTime"`
				Output struct {
					FileURI string `json:"fileUri"`
				} `json:"output"`
				TestCase struct {
					ClassName     string `json:"className"`
					Name          string `json:"name"`
					TestSuiteName string `json:"testSuiteName"`
				} `json:"testCase"`
			} `json:"toolOutputs"`
		} `json:"toolExecution"`
	} `json:"testExecutionStep"`
	ToolExecutionStep struct {
		ToolExecution struct {
			CommandLineArguments []string `json:"commandLineArguments"`
			ExitCode             struct {
				Number int `json:"number"`
			} `json:"exitCode"`
			ToolLogs []struct {
				FileURI string `json:"fileUri"`
			} `json:"toolLogs"`
			ToolOutputs []struct {
				CreationTime struct {
					Nanos   int `json:"nanos"`
					Seconds int `json:"seconds,string"`
				} `json:"creationTime"`
				Output struct {
					FileURI string `json:"fileUri"`
				} `json:"output"`
				TestCase struct {
					ClassName     string `json:"className"`
					Name          string `json:"name"`
					TestSuiteName string `json:"testSuiteName"`
				} `json:"testCase"`
			} `json:"toolOutputs"`
		} `json:"toolExecution"`
	} `json:"toolExecutionStep"`
}

// TestMatrix ...
type TestMatrix struct {
	State          string              `json:"state"` //"TEST_STATE_UNSPECIFIED","VALIDATING","PENDING","RUNNING","FINISHED","ERROR","UNSUPPORTED_ENVIRONMENT","INCOMPATIBLE_ENVIRONMENT","INCOMPATIBLE_ARCHITECTURE","CANCELLED","INVALID"
	Timestamp      string              `json:"timestamp"`
	TestExecutions TestExecutionsModel `json:"testExecutions"`
}

// TestExecutionsModel ...
type TestExecutionsModel []struct {
	State       string `json:"state"`
	Timestamp   string `json:"timestamp"`
	Environment map[string]struct {
		AndroidModelID   string `json:"androidModelId"`
		AndroidVersionID string `json:"androidVersionId"`
		Locale           string `json:"locale"`
		Orientation      string `json:"orientation"`
	} `json:"environment"`
	TestDetails struct {
		ProgressMessages []string `json:"progressMessages"`
	} `json:"testDetails"`
	ToolResultStep struct {
		ProjectID   string `json:"projectId"`
		HistoryID   string `json:"historyId"`
		ExecutionID string `json:"executionId"`
		StepID      string `json:"stepId"`
	} `json:"toolResultsStep"`
	//MetricSamples MetricSampleModel `json:"metricSamples"`
}*/

/*
// TestsList ...
type TestsList struct {
	Steps []struct {
		TestExecutionStep struct {
			TestTiming struct {
				TestProcessDuration struct {
					Seconds string `json:"seconds"`
				} `json:"testProcessDuration"`
			} `json:"testTiming"`
		} `json:"testExecutionStep"`
		State   string `json:"state"`
		Outcome struct {
			Summary string `json:"summary"`
		} `json:"outcome"`
		StepID         string `json:"stepId"`
		DimensionValue []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"dimensionValue"`
	} `json:"steps"`
}

// TestsDetails ...
type TestsDetails struct {
	Steps []struct {
		TestExecutionStep struct {
			TestSuiteOverviews []struct {
				XMLSource struct {
					FileURI string `json:"fileUri"`
				} `json:"xmlSource"`
				TotalCount   int `json:"totalCount"`
				ErrorCount   int `json:"errorCount"`
				FailureCount int `json:"failureCount"`
				SkippedCount int `json:"skippedCount"`
			} `json:"testSuiteOverviews"`
			ToolExecution struct {
				ToolLogs []struct {
					FileURI string `json:"fileUri"`
				} `json:"toolLogs"`
				ToolOutputs []struct {
					Output struct {
						FileURI string `json:"fileUri"`
					} `json:"output"`
				} `json:"toolOutputs"`
			} `json:"toolExecution"`
			TestTiming struct {
				TestProcessDuration struct {
					Seconds int `json:"seconds,string"`
				} `json:"testProcessDuration"`
			} `json:"testTiming"`
		} `json:"testExecutionStep"`
		StepID       string `json:"stepId"`
		CreationTime struct {
			Seconds string `json:"seconds"`
			Nanos   int    `json:"nanos"`
		} `json:"creationTime"`
		CompletionTime struct {
			Seconds string `json:"seconds"`
			Nanos   int    `json:"nanos"`
		} `json:"completionTime"`
		Name        string `json:"name"`
		Description string `json:"description"`
		State       string `json:"state"` //complete,inProgress,pending
		Outcome     struct {
			Summary string `json:"summary"`
		} `json:"outcome"`
		DimensionValue []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"dimensionValue"`
		RunDuration struct {
			Seconds int `json:"seconds,string"`
			Nanos   int `json:"nanos"`
		} `json:"runDuration"`
	} `json:"steps"`
}

// TestMatrixClientResponse ...
type TestMatrixClientResponse struct {
	State          string              `json:"state"` //"TEST_STATE_UNSPECIFIED","VALIDATING","PENDING","RUNNING","FINISHED","ERROR","UNSUPPORTED_ENVIRONMENT","INCOMPATIBLE_ENVIRONMENT","INCOMPATIBLE_ARCHITECTURE","CANCELLED","INVALID"
	TestExecutions TestExecutionsModel `json:"testExecutions"`
	Tests          *TestsList          `json:"tests"`
}
*/
/*
// TestModel ...
type TestModel struct {
	EnvironmentMatrix EnvironmentMatrixModel `json:"environmentMatrix,omitempty"`
	ResultStorage     ResultStorageModel     `json:"resultStorage,omitempty"`
	TestSpecification TestSpecificationModel `json:"testSpecification,omitempty"`
}

// EnvironmentMatrixModel ...
type EnvironmentMatrixModel struct {
	DeviceList *AndroidDevicesList `json:"androidDeviceList,omitempty"`
}

// AndroidDevicesList ...
type AndroidDevicesList struct {
	Devices *[]AndroidDevice `json:"androidDevices,omitempty"`
}

// ResultStorageModel ...
type ResultStorageModel struct {
	GoogleCloudStorage *FileReference `json:"googleCloudStorage,omitempty"`
}

// TestSpecificationModel ...
type TestSpecificationModel struct {
	InstrumentationTest *AndroidInstrumentationTest `json:"androidInstrumentationTest,omitempty"`
	RoboTest            *AndroidRoboTest            `json:"androidRoboTest,omitempty"`
	TestLoop            *AndroidTestLoop            `json:"androidTestLoop,omitempty"`
	TestTimeout         string                      `json:"testTimeout,omitempty"`
}

// FileReference ...
type FileReference struct {
	GcsPath string `json:"gcsPath,omitempty"`
}

// ObbFile ...
type ObbFile struct {
	ObbFileName string         `json:"obbFileName,omitempty"`
	Obb         *FileReference `json:"obb,omitempty"`
}

// AndroidInstrumentationTest ...
type AndroidInstrumentationTest struct {
	AppApk          *FileReference `json:"appApk,omitempty"`
	TestApk         *FileReference `json:"testApk,omitempty"`
	AppPackageID    string         `json:"appPackageId,omitempty"`
	TestPackageID   string         `json:"testPackageId,omitempty"`
	TestRunnerClass string         `json:"testRunnerClass,omitempty"`
	TestTargets     []string       `json:"testTargets,omitempty"`
}

// RoboDirective ...
type RoboDirective struct {
	ResourceName string `json:"resourceName,omitempty"`
	InputText    string `json:"inputText,omitempty"`
	ActionType   string `json:"actionType,omitempty"` //"ACTION_TYPE_UNSPECIFIED","SINGLE_CLICK","ENTER_TEXT"
}

// AndroidRoboTest ...
type AndroidRoboTest struct {
	AppApk             *FileReference   `json:"appApk,omitempty"`
	AppPackageID       string           `json:"appPackageId,omitempty"`
	AppInitialActivity string           `json:"appInitialActivity,omitempty"`
	MaxDepth           int32            `json:"maxDepth,omitempty"`
	MaxSteps           int32            `json:"maxSteps,omitempty"`
	RoboDirectives     *[]RoboDirective `json:"roboDirectives,omitempty"`
}

// AndroidTestLoop ...
type AndroidTestLoop struct {
	AppApk         *FileReference `json:"appApk,omitempty"`
	AppPackageID   string         `json:"appPackageId,omitempty"`
	Scenarios      []int32        `json:"scenarios,omitempty"`
	ScenarioLabels []string       `json:"scenarioLabels,omitempty"`
}

// AndroidDevice ...
type AndroidDevice struct {
	AndroidModelID   string `json:"androidModelId,omitempty"`
	AndroidVersionID string `json:"androidVersionId,omitempty"`
	Locale           string `json:"locale,omitempty"`
	Orientation      string `json:"orientation,omitempty"`
}

// StartTestResponse ...
type StartTestResponse struct {
	Token     string `json:"testMatrixId,omitempty"`
	Timestamp string `json:"timestamp"`
}*/

// TestAsset describes a requested test asset
type TestAsset struct {
	UploadURL string `json:"uploadUrl"`
	GcsPath   string `json:"gcsPath"`
	Filename  string `json:"filename"`
}

// TestAssetsAndroid describes needed Android test asset and is used to return Android test asset upload URLs
type TestAssetsAndroid struct {
	Apk        TestAsset   `json:"apk,omitempty"`
	Aab        TestAsset   `json:"aab,omitmepty"`
	TestApk    TestAsset   `json:"testApk,omitempty"`
	RoboScript TestAsset   `json:"roboScript,omitempty"`
	ObbFiles   []TestAsset `json:"obbFiles,omitempty"`
}

// UploadURLRequest ...
type UploadURLRequest struct {
	AppURL     string `json:"appUrl,omitempty"`
	TestAppURL string `json:"testAppUrl,omitempty"`
}

/*
// DownloadURLResponse ...
type DownloadURLResponse struct {
	Video       string
	XML         string
	Logcat      string
	TestResults string
}

// MetricSampleData ....
type MetricSampleData struct {
	PerfSamples []struct {
		SampleTime struct {
			Seconds int64 `json:"seconds,string"`
			Nanos   int64 `json:"nanos"`
		} `json:"sampleTime"`
		Value float64 `json:"value"`
	} `json:"perfSamples"`
}*/

// MetricSampleModel ...
type MetricSampleModel struct {
	CPU         map[string]float64 `json:"cpu_samples"`
	RAM         map[string]float64 `json:"ram_samples"`
	NetworkDown map[string]float64 `json:"nwd_samples"`
	NetworkUp   map[string]float64 `json:"nwu_samples"`
}

/*
// MatrixSummary ...
type MatrixSummary struct {
	ExecutionID  string `json:"executionId"`
	State        string `json:"state"`
	CreationTime struct {
		Seconds int64 `json:"seconds,string"`
		Nanos   int64 `json:"nanos"`
	} `json:"creationTime"`
	CompletionTime struct {
		Seconds int64 `json:"seconds,string"`
		Nanos   int64 `json:"nanos"`
	} `json:"completionTime"`
	Outcome struct {
		Summary string `json:"summary"`
	} `json:"outcome"`
	TestExecutionMatrixID string `json:"testExecutionMatrixId"`
}
*/
