# Step Template (copy-paste block)

Use inside `milestones/NN-*.md` under ## Steps. Replace `NN.M` and placeholders.

---

### NN.M — Step Title

**Size:** S | M | L  
**Phase:** _A | B | —_  
**Track:** _A | B | —_  
**Depends on:** NN._K_ | M__._K_ | —  
**Blocks:** NN._P_

**Context:** Why this step exists.

**Preconditions:**
- [ ] _
- [ ] Tests pass: `_command_

**Implementation checklist:**
1. _
2. _
3. _

**Files:**

| File | Action |
|---|---|
| `` | Create / Modify / Delete |

**Acceptance criteria:**
- [ ] _
- [ ] _

**Tests (same change set):**
- [ ] `Test_` — _

**Out of scope for this step:**
- _

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| | |

**Rollback:** _

---

## Step quality checklist

Before marking step done:

- [ ] Acceptance criteria are **observable** (not “implement X”)
- [ ] Tests prove behavior, not existence of code
- [ ] Files table matches actual diff
- [ ] Out of scope prevents creep into adjacent milestones
- [ ] If new gap found → [backlog.md](../backlog.md) B-XX, not silent scope add
