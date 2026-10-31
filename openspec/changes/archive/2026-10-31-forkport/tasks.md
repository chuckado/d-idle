# Tasks

```yaml
change: forkport
status: archived
merge: 80ae130
agent: forkport
branch: agent/forkport
worktree: ~/code/d2lod-wt/forkport
skip_specs: true
owns:
  - docs/research/pr148-impact.md
  - docs/BACKLOG.md
exclusive:
  - fork-branch:lab-port-148
gates:
  - "no hashes, hosts or private paths in public files (leakcheck 0)"
result: the lab patch set is ported to the restructured upstream in the private fork; builds, upstream tests as on a pristine build, live smoke and drills D0/D1 on one hidden slot pass. Not switched into any job.
```

## 1. Work

- [x] 1.1 port the lab patch set onto upstream main after #148 (private fork), build, upstream tests
- [x] 1.2 live smoke on one hidden slot (verbs, click modifier, drills D0/D1)
- [x] 1.3 push to the fork's private remote; notes in the private ops repo
- [x] 1.4 anonymous doc lines here
