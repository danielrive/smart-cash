# Commit and Push

Execute a git commit and push to the current branch.

## Before running

1. **Show the current branch** by running `git branch --show-current` and display it clearly to the user.
2. **Warn the user**: "⚠️ All files will be included (git add .)" — make sure they're aware before proceeding.
3. **Request a commit message** from the user if not provided, or use the one they give.

## Commit message rule

Keep the commit message **short and concise**:
- One line summary of the change
- No long descriptions
- Simple but clear (e.g., "Add user login validation", "Fix payment service timeout")

## Steps to execute

1. Display: `Branch: [current-branch-name]`
2. Display: `⚠️ All files will be included (git add .)`
3. Get commit message from user (or use provided message)
4. Run: `git add .`
5. Run: `git commit -m "[commit message]"`
6. Run: `git push origin [current-branch]`

If the user has uncommitted changes, proceed. If they want to review first, show `git status` and let them decide.
