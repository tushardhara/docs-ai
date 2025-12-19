# GitHub Issues - Task Tracking

## Overview
All remaining work has been organized into GitHub issues for systematic development with PR-based workflow.

**Repository**: https://github.com/tushardhara/docs-ai  
**Issues**: https://github.com/tushardhara/docs-ai/issues

---

## Week 3 - Extension Endpoint Refinements (Optional)

### Issue #1: Fine-tune DOM parsing for edge cases
**Priority**: Low (optional)  
**Link**: https://github.com/tushardhara/docs-ai/issues/1  
**Tasks**: 
- Test with complex nested DOM structures
- Handle iframes and shadow DOM
- Validate selector extraction
- Add error handling for malformed DOM data

### Issue #2: Optimize hybrid search for relevance
**Priority**: Medium  
**Link**: https://github.com/tushardhara/docs-ai/issues/2  
**Tasks**:
- Analyze search results quality
- Adjust pgvector vs Meilisearch weights
- Implement result deduplication
- Add relevance scoring metrics

---

## Week 4 - Browser Extension UI (Critical Path)

### Issue #3: Set up Chrome extension directory structure 🔴
**Priority**: HIGH - Blocks all other Week 4 work  
**Link**: https://github.com/tushardhara/docs-ai/issues/3  
**Tasks**:
- Create `extension/` directory structure
- Set up manifest.json (Chrome v3)
- Add icons and assets
- Configure permissions

### Issue #4: Configure TypeScript + React for extension 🔴
**Priority**: HIGH  
**Link**: https://github.com/tushardhara/docs-ai/issues/4  
**Dependencies**: Requires #3  
**Tasks**:
- Initialize package.json
- Install React, TypeScript, @types/chrome
- Configure tsconfig.json
- Set up webpack/esbuild

### Issue #5: Implement DOM capture content script 🔴
**Priority**: HIGH - Core functionality  
**Link**: https://github.com/tushardhara/docs-ai/issues/5  
**Tasks**:
- Create `content/capture.ts`
- Implement DOM snapshot capture
- Implement screenshot capture
- Extract interactive elements
- Add message passing

### Issue #6: Build popup UI with React 🔴
**Priority**: HIGH - Main user interface  
**Link**: https://github.com/tushardhara/docs-ai/issues/6  
**Tasks**:
- Create `popup/App.tsx`
- Implement chat interface
- Add "Analyze Page" button
- Display guidance steps
- Add settings panel
- Style with CSS

### Issue #7: Create API client utility 🔴
**Priority**: HIGH  
**Link**: https://github.com/tushardhara/docs-ai/issues/7  
**Tasks**:
- Create `utils/api.ts`
- Implement `chatWithExtension()` API call
- Add TypeScript interfaces
- Handle authentication
- Implement error handling

### Issue #8: Implement local storage management
**Priority**: Medium  
**Link**: https://github.com/tushardhara/docs-ai/issues/8  
**Tasks**:
- Create `utils/storage.ts`
- Save/load settings (API endpoint, project ID)
- Save/load chat history
- Add data migration support

### Issue #9: Add element highlighting functionality
**Priority**: Medium  
**Link**: https://github.com/tushardhara/docs-ai/issues/9  
**Tasks**:
- Create `content/highlighter.ts`
- Implement `highlightElement(selector)`
- Add scroll-to-element functionality
- Support multiple highlights
- Test with various layouts

### Issue #10: Test extension in Chrome developer mode 🔴
**Priority**: HIGH - Quality assurance  
**Link**: https://github.com/tushardhara/docs-ai/issues/10  
**Tasks**:
- Load unpacked extension
- Test on 5+ SaaS applications
- Verify DOM capture, screenshots, API calls
- Test element highlighting
- Check performance

### Issue #11: End-to-end extension testing 🔴
**Priority**: HIGH - MVP validation  
**Link**: https://github.com/tushardhara/docs-ai/issues/11  
**Tasks**:
- Test 8 complete user scenarios
- Verify full setup flow works
- Test across multiple tabs
- Test error handling
- Validate chat history persistence

---

## Production Readiness

### Issue #12: Set up real API keys for media processing
**Priority**: Medium  
**Link**: https://github.com/tushardhara/docs-ai/issues/12  
**Tasks**:
- Google Cloud Vision API (OCR)
- YouTube Data API v3
- Whisper/AssemblyAI API
- Update `.env.example`
- Test with real API calls

### Issue #13: Configure deployment and monitoring
**Priority**: Medium  
**Link**: https://github.com/tushardhara/docs-ai/issues/13  
**Tasks**:
- Production Docker Compose config
- SSL certificates setup
- Health check monitoring
- Error tracking (Sentry)
- PostgreSQL backup strategy
- CI/CD pipeline setup

---

## Development Workflow

### How to Work on an Issue

1. **Pick an issue** from https://github.com/tushardhara/docs-ai/issues
2. **Create feature branch**: `git checkout -b feature/issue-N-description`
3. **Make changes** following project patterns
4. **Test**: `go build ./...` and `go test ./...`
5. **Commit**: `git commit -m "feat: description (#N)"`
6. **Push**: `git push origin feature/issue-N-description`
7. **Create PR** linking to the issue with `Fixes #N`
8. **Get review** and approval
9. **Merge** (squash and merge recommended)
10. **Issue auto-closes** via PR merge

### Commit Convention
- `feat:` - New features (#5, #6, #7)
- `fix:` - Bug fixes (#1, #2)
- `docs:` - Documentation updates
- `test:` - Test improvements (#10, #11)
- `chore:` - Maintenance (#13)

### Branch Protection
⚠️ **NO direct pushes to `main`** - All changes must go through PRs

---

## Recommended Order for Week 4

### Phase 1: Foundation (Days 1-2)
1. **#3** - Directory structure ⬅️ Start here
2. **#4** - TypeScript + React setup
3. **#7** - API client utility

### Phase 2: Core Features (Days 3-4)
4. **#5** - DOM capture content script
5. **#6** - Popup UI with React
6. **#8** - Local storage management

### Phase 3: Polish & Testing (Days 5-6)
7. **#9** - Element highlighting
8. **#10** - Chrome developer testing
9. **#11** - End-to-end testing

### Optional Refinements (Anytime)
- **#1** - DOM parsing edge cases
- **#2** - Search relevance optimization

### Production (After MVP)
- **#12** - Real API keys
- **#13** - Deployment & monitoring

---

## Current Status Summary

**Completed**: 
- ✅ Week 1 - Database foundation
- ✅ Week 2 - Media handlers (OCR, YouTube, Video)
- ✅ Week 3 - Extension endpoint (~95%)

**In Progress**: 
- ⏳ Week 4 - Browser extension UI (0% - ready to start)

**Total Issues Created**: 13  
**High Priority Issues**: 6 (#3, #4, #5, #6, #7, #10, #11)  
**Medium Priority Issues**: 4 (#2, #8, #9, #12, #13)  
**Low Priority Issues**: 1 (#1)

---

## Resources

- **Workflow Guide**: `docs/GITHUB_WORKFLOW.md`
- **Development Patterns**: `.github/copilot-instructions.md`
- **Quick Start**: `docs/START_HERE.md`
- **Commands Reference**: `docs/COMMANDS.md`
- **Project Status**: `docs/COMPLETION_STATUS.md`
