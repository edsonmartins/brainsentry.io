export interface FaqItem {
  q: string;
  a: string;
}

export const faqItems: FaqItem[] = [
  {
    q: "How is Brain Sentry different from RAG?",
    a: "RAG is read-only retrieval. Brain Sentry is full Agent Memory with read-write, learning, and graph relationships. Think of RAG as a filing cabinet. Brain Sentry is a second brain that learns and grows.",
  },
  {
    q: "Does it work with my AI coding assistant?",
    a: "Yes! Works with Claude Code (Anthropic), Cursor, GitHub Copilot, Continue, and any LLM-based coding assistant.",
  },
  {
    q: "Do I need to change my workflow?",
    a: "No! Brain Sentry operates transparently through autonomous interception. Your workflow stays the same.",
  },
  {
    q: "Is my code data secure?",
    a: "Absolutely: Self-hosted deployment (your infrastructure), no data leaves your servers, full audit trail, encryption at rest and in transit, open source (audit the code yourself).",
  },
  {
    q: "What's the performance overhead?",
    a: "Minimal: Latency p95: <500ms, p99: <1000ms, Throughput: >100 req/sec (single instance).",
  },
  {
    q: "Can I use it with my team?",
    a: "Yes! Multi-user support with shared memory pool, role-based access control (Phase 4), personal memories (private), and audit logs.",
  },
  {
    q: "Is there a hosted/SaaS version?",
    a: "Coming Q3 2025: Q1-Q2 self-hosted open source, Q3 SaaS beta (waitlist), Q4 SaaS general availability. Self-hosted remains free forever.",
  },
  {
    q: "How do I get started?",
    a: "Three paths: 1) Join Waitlist → Early access + guided onboarding, 2) Install Now → Follow QUICK_START.md (30 min), 3) Contribute → Clone repo, read CONTRIBUTING.md.",
  },
];
