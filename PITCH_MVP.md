# CGAP — AI Browser Extension for SaaS (MVP Pitch)

## One-Liner
**Click CGAP on any SaaS dashboard → ask "How do I create X?" → get step-by-step guidance with clickable selectors grounded in your docs + videos.**

## Problem
1. **SaaS complexity**: Users get stuck on basic workflows (Mixpanel: create dashboard; Stripe: set up payment; HubSpot: build workflow)
2. **Support overload**: Teams answer the same 20 questions 1000+ times per month
3. **Generic AI fails**: ChatGPT + browser agents hallucinate; they don't know YOUR specific UI
4. **Docs alone don't work**: PDFs and videos don't guide users through actual clicks

## Solution
**CGAP Browser Extension** (Install on Chrome/Firefox):

```
1. User opens Mixpanel, logs in
2. Clicks CGAP extension icon
3. Types: "How do I create a dashboard?"
4. Gets guidance:
   - "Step 1: Click Dashboards nav link (selector: .nav-dashboards)"
   - "Step 2: Click New Dashboard button"
   - "Step 3: Enter name 'My Dashboard'"
   - "Step 4: Click Save"
   - [Optional] Auto-click each step with user confirmation
```

## Why It Works
- **Knows YOUR dashboard**: Trained on your docs + OCR of screenshots + YouTube transcripts
- **No hallucinations**: Every step has actual UI selector (e.g., `.create-btn`)
- **Works for any SaaS**: Mixpanel, Stripe, HubSpot, Salesforce, Shopify, etc.
- **Beats generic agents**: Scoped knowledge = higher accuracy + lower cost + defensible
- **Reduces support**: 20-40% ticket deflection rate (measured by Kapa)

## Market Position
| Product | Type | Approach | Limitation |
|---------|------|----------|-----------|
| Stagehand/AgentQL | General browser agent | Works on any website | Expensive, hallucination-prone |
| ChatGPT + Browser | Generic LLM + agent | Any website | No product knowledge, slow |
| Kapa | Documentation Q&A | Text-only | Doesn't guide actual clicks |
| **CGAP (MVP)** | **Browser extension** | **Scoped to one SaaS** | **Can't scale across 1000 SaaS** |

Our advantage: **For ONE product, we're 10x better than Stagehand.**

## MVP Demo (End of Week 4)
```
Live Demo:
1. Open Mixpanel dashboard in Chrome
2. Click CGAP extension
3. Ask: "Create a dashboard for user behavior"
4. Extension shows:
   - DOM entities (buttons, inputs detected)
   - Guidance: "Go to Dashboards → New → Enter name → Save"
   - Optional: Auto-click Dashboards button
5. Accuracy: 90%+ (vs. ChatGPT: 30% on SaaS-specific tasks)
```

## Metrics (Pre-launch)
- Extension loads: ✅ (Dev mode)
- DOM capture: ✅ (Accurate)
- API latency: <1s
- Guidance accuracy: >85%
- Demo videos: 3+ (Mixpanel, Stripe, HubSpot)

## GTM (Month 2+)
- **Pilot**: 3 customers at $5k MRR each
- **Measure**: Support ticket reduction (-30%), user satisfaction (NPS >50)
- **Productize**: Self-service customer onboarding
- **Monetize**: $99–499/month per SaaS customer (based on support volume)

## Why Now
- Browser AI is hot (YC funding: Stagehand, AgentQL, browser-use)
- BUT: Generic agents don't work well for product-specific UIs
- Gap: Need something **scoped + knowledge-grounded**
- Timing: Support costs rising; Kapa raising Series B at $500M+ valuation

## Why We Win
1. **Speed**: 4-week MVP vs. 6-month enterprise product
2. **Accuracy**: Scoped knowledge > generic agents (90% vs. 30%)
3. **Cost**: Extension (free hosting) vs. Stagehand ($$ for browser fleet)
4. **Moat**: Knowledge of customer's specific SaaS + their docs = defensible

## Ask
- **Series Seed**: $2M for product, pilot customers, GTM
- **Use of funds**: 
  - Product (extension + API): 40%
  - Customer pilot + success: 30%
  - Team (2 eng, 1 product): 20%
  - Operations: 10%

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Browser extension approval slow | Publish day 1 in dev mode; users trust CGAP over random Chrome store |
| Generic agents improve | Our advantage is knowledge + scope; defensible for years |
| Customer acquisition | Direct sales to support teams (Mixpanel, Stripe users); prove ROI |
| LLM quality | Proprietary prompts + customer-specific fine-tuning |

## Team
- **Tushar** (Founder): 5y SaaS, AI infrastructure
- (Hiring): Product, Sales, Customer Success
