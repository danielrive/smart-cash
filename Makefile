.PHONY: help workflows list-runs watch-run

help:
	@echo "Smart-Cash Workflow Management"
	@echo ""
	@echo "Available targets:"
	@echo "  help              Show this help message"
	@echo "  workflows         Show all available workflows"
	@echo "  list-runs         List recent workflow runs"
	@echo "  watch-run         Watch the latest workflow run"
	@echo ""
	@echo "To run workflows, use the ./run-workflow.sh script:"
	@echo "  ./run-workflow.sh --help"
	@echo ""
	@echo "Examples:"
	@echo "  make workflows                                    # List all workflows"
	@echo "  make list-runs                                    # Show recent runs"
	@echo "  make watch-run                                    # Watch latest run"
	@echo "  ./run-workflow.sh infra-develop --stage 1-base-stage"
	@echo "  ./run-workflow.sh service-develop --service user-service"

workflows:
	@echo "Fetching available workflows..."
	@gh workflow list 2>/dev/null || echo "Error: gh CLI not authenticated. Run: gh auth login"

list-runs:
	@echo "Recent workflow runs:"
	@gh run list --limit 10 2>/dev/null || echo "Error: gh CLI not authenticated. Run: gh auth login"

watch-run:
	@echo "Watching latest workflow run..."
	@gh run watch 2>/dev/null || echo "Error: gh CLI not authenticated or no runs available. Run: gh auth login"
