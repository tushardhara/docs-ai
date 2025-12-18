# CGAP MVP - 1-Month Sprint Planning

## 📚 Files to Read (In This Order)

### 1. **Start Here** (5 min read)
- 📄 [MVP_SUMMARY.md](./MVP_SUMMARY.md) - Overview of 4-week sprint
- 🎯 [WHAT_TO_BUILD.md](./WHAT_TO_BUILD.md) - What's built vs what's new

### 2. **Strategy** (10 min read)
- 🚀 [PITCH_MVP.md](./PITCH_MVP.md) - Customer pitch + market positioning
- 📋 [TODO_MVP.md](./TODO_MVP.md) - Week-by-week checklist

### 3. **Technical** (20 min read)
- 🛠️ [README_MVP.md](./README_MVP.md) - How to build + run everything
- 📖 [PRD.md](./PRD.md) - Phase 0 requirements (Phase 1-8 deferred)

### 4. **Reference** (Throughout)
- 🏗️ [README.md](./README.md) - Original docs (partially updated)
- 📋 [TODO.md](./TODO.md) - Old roadmap (now focused on Week 1-4)

---

## 🎬 Quick Start (2 minutes)

```bash
# 1. Clone & setup
cd /Users/tushar.dhara/docs-ai
docker-compose up

# 2. In Terminal 2: Backend
go run cmd/api/main.go
go run cmd/worker/main.go

# 3. In Terminal 3: Extension
cd extension
npm install
npm run build

# 4. Load extension
# Chrome: Settings → Extensions → Load unpacked → extension/dist/

# 5. Test
curl http://localhost:8080/health
```

---

## 📅 4-Week Sprint Breakdown

| Week | Focus | Deliverable | Done When |
|------|-------|-------------|-----------|
| **1** | Database | Schema ready | `go build ./...` passes |
| **2** | Media | OCR + YouTube working | Ingest image → text stored |
| **3** | Endpoint | Extension chat API | POST /v1/extension/chat works |
| **4** | Extension | Chrome plugin | Demo works on Mixpanel |

---

## 🚀 Week 1 Tasks (Mon-Fri)

### Daily Standup Template
```
Monday:
- [ ] Create migrations/ SQL files
- [ ] Run goose up
- [ ] Schema created in Postgres

Tuesday:
- [ ] Update ingest handler to record source_id
- [ ] Test: POST /v1/ingest → documents have source_id
- [ ] Build passes

Wednesday:
- [ ] Create media_items table
- [ ] Create extracted_text table
- [ ] Create document_sources mapping

Thursday:
- [ ] Update worker to handle media_items
- [ ] Test basic flow
- [ ] All migrations clean

Friday:
- [ ] Code review
- [ ] All tests passing
- [ ] Document in README
- [ ] Ready for Week 2 (media handlers)
```

---

## 📊 Success Metrics (After Week 4)

| Metric | Target | Status |
|--------|--------|--------|
| Build status | ✅ PASS | `go build ./...` |
| Tests passing | ✅ >70% | `go test ./...` |
| Extension loads | ✅ YES | Chrome dev mode |
| Demo works | ✅ YES | Mixpanel → ask → get steps |
| Latency | ✅ <1s | API response time |
| Accuracy | ✅ >85% | Manual testing |

---

## 🎯 Three Most Important Files

### 1. [WHAT_TO_BUILD.md](./WHAT_TO_BUILD.md)
**Read first.** Tells you what's done vs what you need to build. Prevents confusion.

### 2. [TODO_MVP.md](./TODO_MVP.md)
**Reference daily.** Week-by-week checklist. Check off as you go.

### 3. [README_MVP.md](./README_MVP.md)
**How-to guide.** Build + run + test commands.

---

## 💡 Key Insights

1. **Don't build the old Kapa competitor.** Build the browser extension first.
2. **Focus on ONE SaaS dashboard.** Make it perfect for Mixpanel, then expand.
3. **Demo > perfection.** Make it work, show a customer, iterate.
4. **Reuse existing code.** API, worker, search, LLM are already built. Just add extension.

---

## 🚨 Common Pitfalls to Avoid

| Pitfall | How to Avoid |
|---------|-------------|
| Trying to support ASR (Whisper) in Week 2 | Just do OCR + YouTube. ASR is Phase 2. |
| Building complex DOM selector logic | Use simple CSS selectors + regex. Refine later. |
| Over-engineering extension UI | Just show steps as numbered list. Animation is Phase 2. |
| Trying to do Phases 1-8 | Focus only on Week 1-4 checklist. Everything else is "later". |
| Waiting for perfect code | Ship working MVP even if messy. Polish in Phase 1. |

---

## 📞 Daily Resources

### If you get stuck:
1. Check [TODO_MVP.md](./TODO_MVP.md) - detailed step-by-step checklist
2. Check [README_MVP.md](./README_MVP.md) - code examples + commands
3. Check [WHAT_TO_BUILD.md](./WHAT_TO_BUILD.md) - dependency explanation
4. Google it (you're doing standard stuff: OCR, YouTube API, Chrome extension)

### If you're ahead of schedule:
1. Add error handling (Week 3)
2. Add test coverage (any week)
3. Improve LLM prompt (Week 3)
4. Build demo video (Week 4)

---

## 🎬 Demo Day (Friday of Week 4)

### Before Demo:
- [ ] Docs ingested into Mixpanel project
- [ ] Extension loaded on Chrome
- [ ] API running on :8080
- [ ] Test once: Works end-to-end

### During Demo (3 min):
1. Open Mixpanel.com (logged in)
2. Click CGAP extension
3. Ask: "How do I create a dashboard?"
4. See: Guidance with numbered steps
5. Optionally: Show auto-click feature (don't actually click)

### After Demo:
- Share with first customer
- Get feedback
- Iterate → Phase 1

---

## 📈 What Success Looks Like (Week 4 Friday)

✅ **Technical** (Code works)
- Extension loads without errors
- API endpoint responds <1s
- Can ask questions on real SaaS dashboard
- Demo completes without issues

✅ **Business** (Ready to sell)
- Story is clear (browser extension MVP)
- Demo is impressive (live on Mixpanel)
- Customer can sign up and try
- Support team can measure deflection

✅ **Execution** (Team aligned)
- Code is on `main` branch
- README explains how to use
- Checklist is 100% complete
- Team ready for customer pilot

---

## 🏁 One Month From Now

**Friday, January 17, 2025**

You will have:
1. ✅ Working browser extension
2. ✅ Running on Chrome
3. ✅ Live demo on Mixpanel/Stripe/HubSpot
4. ✅ Ready for first customer pilot
5. ✅ Pitch ready for investors

**Status: MVP Complete → Customer Pilot Ready**

---

## 📞 Questions?

- **"What do I build first?"** → Week 1: Database migrations
- **"How long will this take?"** → 4 weeks (1 month exactly)
- **"What if I get stuck?"** → Read WHAT_TO_BUILD.md + README_MVP.md
- **"Can I parallelize the work?"** → Yes, split into: Backend (Week 1-3) + Frontend (Week 4)
- **"What happens after MVP?"** → Phase 1: Auto-click, analytics, deflection refinement

---

Start with [MVP_SUMMARY.md](./MVP_SUMMARY.md). ✅
