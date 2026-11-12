# Master-Prompt.md — Goal-Driven Task Template

Use this template whenever you assign work to an AI coding agent. It helps you describe **what problem to solve, why it matters, and what done looks like**, and leaves the agent to work out how.

Your desired outcome is the source of truth. Any implementation idea you include is a hypothesis for the agent to test, not an order.

## How This Fits With Agent.md

`Agent.md` holds the standing principles for every task. This template holds the specifics of one task. Keep them separate: do not repeat global principles in each prompt, and do not put task details in `Agent.md`.

| Dimension | Agent.md (global principles) | This prompt (task instructions) |
|---|---|---|
| Architecture | Establishes direction | States what this task must fit into or preserve |
| Security | Sets expectations | States the specific rules to enforce here |
| Scalability | Sets design goals | States the actual load, volume, and limits |
| Maintainability | Establishes standards | States the quality bar for this change |
| Functionality | Limited guidance | Defines the exact behavior required |

## The Template

Copy the block below and fill it in. Write plain sentences. Leave a section out if it truly does not apply, but always keep **Goal**, **Problem Statement**, and **Acceptance Criteria**.

```markdown
## Goal
<What should be achieved, in one or two sentences. Describe the result, not the work.>

## Problem Statement
<What happens today. What is wrong, missing, slow, or risky. Include evidence:
error messages, reproduction steps, numbers, examples. Mention what has
already been tried.>

## Business Context
<Why this matters. Who is affected and how. Urgency, and the cost of leaving it.>

## Existing Behavior
<How the system works today in the affected area, as far as you know.
Mark what you know for certain and what you are guessing.>

## Constraints
<What must be preserved or avoided, and the reason for each: technical,
business, security, compliance, compatibility, performance, deadlines,
parts of the system that are off-limits.>

## Non-Goals
<What this task should not attempt, even if it looks related.>

## Priorities
<When trade-offs appear, what wins? For example: correctness over speed of
delivery, a small diff over ideal structure, no downtime over simplicity.>

## Quality Bar
<Keep the one line that applies and delete the others. Add any task-specific
standards, such as a coverage threshold or a compliance rule.>
- Prototype: works on the main path; minimal error handling; tests optional; shortcuts labelled.
- Production: handles invalid input and failures; tests for new behavior and regressions; logging where failures would be hidden; follows project conventions.
- Critical: everything in Production, plus concurrency and partial-failure safety, authorization checked and tested, a rollback path, performance verified at realistic volume, and an audit trail where required.

## Expected Outcome
<What should be true once the work is complete, from the point of view of
users and of the system.>

## Acceptance Criteria
<Observable, checkable conditions that prove success.>
- [ ] ...
- [ ] ...

## Verification
<How the agent can check its own work: how to reproduce the problem, which
commands, environments, data, or logs it can use, and what it cannot access.>

## Ideas and Hints (optional)
<Suspicions, suggested approaches, places to look. The agent will treat these
as hypotheses and may recommend something different.>

## Mode and Autonomy (optional)
<Choose one: investigate and report only / propose an approach and wait /
implement end to end. Say what needs your approval first.
Default: implement end to end, confirming first only for high-impact or
irreversible changes.>
```

## Standalone Working Instructions

If `Agent.md` is loaded, skip this section. If the agent has no standing instructions, append this block to your prompt.

```markdown
## How to Work

1. **Understand the problem.** Restate it briefly. Reason from first principles about what must be true for the outcome to hold.
2. **Explore the codebase.** Read the relevant code, tests, and conventions before changing anything. Think about the whole system: callers, consumers, data, and dependencies.
3. **Investigate the root cause.** Confirm it with evidence. Do not stop at the symptom.
4. **Discover unstated requirements.** Identify edge cases, failure modes, security implications, and operational needs I did not mention. Tell me about the ones that change the solution.
5. **Evaluate possible solutions.** Compare the realistic options, including any I suggested. Challenge my assumptions where the evidence contradicts them.
6. **Recommend an approach with reasoning.** Choose the simplest solution that fully solves the problem. For high-impact or irreversible changes, confirm with me before implementing. Otherwise proceed.
7. **Implement.** Make a minimal, maintainable change that follows existing patterns and preserves existing behavior unless this task requires otherwise.
8. **Validate.** Run the relevant tests and checks. Verify each acceptance criterion. Never report a test as passing unless you ran it.
9. **Report.** Give the outcome, root cause, approach and reasoning, changes, verification results, assumptions, trade-offs, and remaining risks.

Scale the depth of all this to the size of the task. Ask me questions only when the missing information would materially change the solution.
```

## Short Form for Small Tasks

Small, low-risk work does not need the full template. Three lines are enough:

```markdown
Goal: <the result you want>
Problem: <what is wrong today, with evidence>
Done when: <how you will know>
Quality bar: <Prototype / Production / Critical; Production if omitted>
```

## Quality Bar

The quality bar tells the agent how much rigor this task deserves. It replaces vague labels such as "enterprise-grade" with expectations the agent can meet and you can check. If you leave it out, the agent should assume **Production**.

| Expectation | Prototype | Production | Critical |
|---|---|---|---|
| Typical use | Spikes, demos, throwaway scripts | Normal product work | Money, auth, personal data, migrations, shared platform code |
| Error handling | Main path only | Invalid input and expected failures handled; no silent failures | Also partial failures, retries, duplicates, and concurrency |
| Tests | Optional | New behavior and regressions covered | Also failure paths, authorization, and boundary cases |
| Security | No secrets in code | Input validated; authorization at the point of access | Also threat cases considered and tested; least privilege verified |
| Performance | Not a concern | No obvious waste at realistic scale | Measured at realistic volume against a stated target |
| Operations | None | Logging where a failure would be hidden | Also metrics or alerts, a rollback path, and safe migrations |
| Documentation | Shortcuts labelled | Non-obvious decisions explained | Also operational notes and an audit trail where required |
| Compatibility | May break | Preserved for public interfaces and stored data | Preserved, with a tested migration path for any break |

Adjust these definitions to your team's standards, then keep them stable so the three names mean the same thing in every prompt.

The quality bar sets how rigorous the result must be. It does not set how the agent builds it, and a higher bar is not a request for more abstraction. A Critical change should still be the simplest one that meets the Critical expectations.

## Writing Each Section Well

- **Goal.** If your goal names a file, function, or library, you have written a solution. Ask yourself why you want that change, and write down the answer.
- **Problem Statement.** Evidence is worth more than interpretation. Paste the error, the slow query, the confusing screen, the failing input.
- **Business Context.** This is how the agent makes trade-offs the way you would. "Finance closes the books with this report on the first of the month" changes the engineering decisions.
- **Existing Behavior.** Share what you know and admit what you do not. The agent will verify it against the code.
- **Constraints.** Be explicit here. This is the right place for firm instructions: what must not break, what must not be touched, what standards apply. Give the reason for each one. A constraint with a reason can be applied sensibly to cases you did not foresee. If a specific technology or approach is truly mandatory, state it here as a constraint, with the reason.
- **Non-Goals.** Capable agents notice related problems. Naming what is out of scope keeps the change focused and reviewable.
- **Priorities.** Every real task involves trade-offs. Stating which side wins lets the agent decide the way you would, without asking.
- **Quality Bar.** Choose by what a failure would cost, not by how important the task feels. Most work is Production. Reserve Critical for changes where a mistake is expensive or hard to undo.
- **Verification.** An agent that can reproduce the problem and run the checks will correct its own mistakes. One that cannot is guessing.
- **Mode and Autonomy.** Use "investigate and report only" when you do not yet trust your own problem statement. A diagnosis is often worth having before any code changes.
- **Expected Outcome and Acceptance Criteria.** The outcome describes the end state. The criteria are how someone checks it. Good criteria are observable: a behavior, a number, a passing test.

## Additions by Task Type

Add the relevant points to the sections above.

| Task type | Add to your prompt |
|---|---|
| **New feature** | Who uses it and the main scenarios. What is explicitly out of scope. How it should behave for empty, invalid, and unauthorized cases, if you know. |
| **Bug fix** | Reproduction steps, expected versus actual behavior, when it started, how often, who is affected. Ask for a regression test. |
| **Refactoring** | The pain the current structure causes. What behavior must remain identical. How equivalence will be proven. |
| **Performance optimization** | Current measurements, the target, realistic data volumes and load, and how to measure. What may not be traded away, such as correctness or freshness. |
| **Architecture improvement** | The limitation being hit and the change it blocks. Known future needs, with a clear line between confirmed and speculative. Appetite for migration effort. |
| **Technical debt** | What the debt costs today: incidents, slow delivery, onboarding. The boundary of the cleanup, and what should be left alone. |
| **System integration** | The external system's contract, limits, and failure behavior. Authentication, data ownership, retries, idempotency, and what happens when the other side is down. |
| **Complex multi-step work** | The end goal and any required ordering. Where you want checkpoints. What must keep working between steps. Ask for a short plan before implementation. |

## Solution-Oriented Versus Goal-Oriented Prompts

### Bug fix

**Solution-oriented:**

> In `OrderService.sendConfirmation()`, add a check so the email is not sent if `order.emailSent` is true.

**Goal-oriented:**

> **Goal:** Each customer receives exactly one confirmation email per order.
> **Problem:** About 2% of customers receive two or three confirmation emails. It occurs mostly at peak traffic. Support receives roughly 30 complaints a week.
> **Constraints:** Customers must never receive zero emails. The email provider must stay the same.
> **Acceptance criteria:** Duplicates cannot occur under concurrent or retried processing. A test reproduces the original failure and now passes.

The first prompt gets you a flag check. The second lets the agent find out that payment webhooks are retried and processed concurrently, where a flag check alone would still race.

### Performance

**Solution-oriented:**

> Add Redis caching to the `/dashboard` endpoint.

**Goal-oriented:**

> **Goal:** The dashboard loads in under one second for our largest accounts.
> **Problem:** Accounts with more than 5,000 projects wait 8 to 12 seconds. Smaller accounts are fine.
> **Constraints:** Figures shown must be no more than a minute old. Prefer no new infrastructure unless it is clearly justified.
> **Acceptance criteria:** p95 under one second with 10,000 projects, measured before and after.

The agent may find a query running once per project. Fixing that removes the problem, and a cache would only have hidden it.

### New feature

**Solution-oriented:**

> Create a `GET /api/export` endpoint that returns the users table as CSV using the csv library.

**Goal-oriented:**

> **Goal:** Account administrators can get their user list into a spreadsheet for quarterly access reviews.
> **Context:** Compliance teams at enterprise customers do this review every quarter and currently copy the data by hand.
> **Constraints:** Only administrators of the account may export, and only that account's users. Some accounts have more than 100,000 users.
> **Acceptance criteria:** An administrator can export all users with role and last-login date. A non-administrator cannot. A large export completes without timing out.

The second version brings authorization, tenant isolation, and scale into the design from the start.

### Refactoring

**Solution-oriented:**

> Split `PaymentProcessor` into three classes using the strategy pattern.

**Goal-oriented:**

> **Goal:** Adding a payment provider should not require touching code for the existing providers.
> **Problem:** The last two provider additions each caused a regression in an unrelated provider, because all provider logic shares one 1,400-line class.
> **Constraints:** No behavior change for existing providers. No changes to the public interface used by checkout.
> **Acceptance criteria:** Existing tests pass unchanged. A new provider can be added by adding code, not by editing shared logic.

## Goal-Oriented Does Not Mean Vague

Removing the solution from a prompt is only half the work. The agent still needs evidence, limits, and a way to recognize success.

| Style | Prompt | What goes wrong |
|---|---|---|
| Too prescriptive | "Add an index on `orders.customer_id`." | The agent adds it, whether or not it is the cause. |
| Too vague | "Make the orders page faster." | The agent must guess how fast, for whom, and at what cost. |
| Goal-oriented | "The orders page takes six seconds for customers with more than 2,000 orders. It should load in under one second, and what the page shows must not change." | Nothing. The agent knows the problem, the target, and the limit, and is free to find the cause. |

## Before You Send

- Does the goal describe a result, not an implementation?
- Would someone unfamiliar with the task understand why it matters?
- Are your constraints explicit, especially what must not change?
- Can each acceptance criterion be checked?
- Are your own solution ideas marked as ideas?
- Have you said what is out of scope and which trade-offs win?
- Does the quality bar match what a failure would cost?
- Can the agent reproduce the problem and verify the result with what you gave it?

## Your Role After Sending

Read the agent's root-cause finding and reasoning before you read its code. Confirm the assumptions it lists. Check the verification evidence against your acceptance criteria. You remain responsible for the problem definition and for the quality of what ships.
