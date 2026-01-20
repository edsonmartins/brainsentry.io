export interface UseCase {
  id: string;
  title: string;
  challenge: string;
  solution: string;
  result: string;
  quote: string;
  author: string;
}

export const useCases: UseCase[] = [
  {
    id: "team-consistency",
    title: "Team Consistency",
    challenge: "Our 10-person team uses Claude Code. Everyone codes differently, and code review is a bottleneck.",
    solution: "Brain Sentry stores team patterns as Procedural memories. When any developer uses their AI assistant, it automatically follows established conventions.",
    result: "40% reduction in code review feedback",
    quote: "Brain Sentry acts like a senior developer looking over everyone's shoulder—without the awkwardness.",
    author: "Tech Lead, Series B Startup",
  },
  {
    id: "onboarding",
    title: "Onboarding",
    challenge: "New developers spend weeks reading docs before they're productive. By then, docs are outdated.",
    solution: "Brain Sentry captures Episodic memories of decisions and Semantic memories of architecture. New team members' AI assistants have instant access to tribal knowledge.",
    result: "50% faster onboarding time",
    quote: "Our junior was pushing prod code in week one. Brain Sentry gave them context we didn't even document.",
    author: "Engineering Manager, Enterprise SaaS",
  },
  {
    id: "bug-tracking",
    title: "Bug Tracking",
    challenge: "We fix the same bugs repeatedly. Knowledge is in Jira tickets, but nobody searches old tickets when debugging.",
    solution: "Brain Sentry's Hindsight Notes capture bug resolutions with full context. When similar errors appear, AI automatically surfaces past fixes.",
    result: "30% faster debugging",
    quote: "It's like having a developer with perfect memory of every bug we've ever fixed.",
    author: "Senior Developer, Open Source",
  },
  {
    id: "refactoring",
    title: "Refactoring",
    challenge: "Large refactors break things. We spend more time fixing ripple effects than improving code.",
    solution: "Graph relationships track dependencies automatically. Before refactoring: 'What depends on this?' Get comprehensive impact analysis.",
    result: "Zero breaking changes in last major refactor",
    quote: "The graph showed us dependencies we didn't know existed. Saved us from disaster.",
    author: "Staff Engineer, Fintech",
  },
];
