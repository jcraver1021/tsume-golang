# Book Citation Analysis Project

This directory contains tools and documentation for systematically evaluating the factuality and quality of citations in the book.

---

## Project Structure

```
proj/debunk/toxic_sociopathy/
├── README.md                    # This file - project overview
├── citations.md                 # Complete citation list with full text
├── evaluation_schema.md         # Detailed evaluation methodology
├── quick_reference.md           # Quick reference for evaluators
├── evaluation_template.yaml     # Blank template for evaluations
├── evaluation_example.yaml      # Fully worked example
└── evaluations/                 # Your completed evaluations go here
    ├── lie01/                   # Organize by section
    ├── lie02/
    ├── lie03/
    ├── lie04/
    └── lie05/
```

---

## Quick Start

### 1. Download Citations

First, use the labrador tool to download all accessible sources:

```bash
cd /Users/jcraver/workbench/tsume-golang

# Conservative settings (recommended first run)
./cmd/labrador/labrador \
  -file citations.yaml \
  -worker-count 2 \
  -retry-count 3 \
  -backoff 2000 \
  -output-dir citations_downloaded
```

See `LABRADOR_USAGE.md` for detailed download instructions.

### 2. Set Up Evaluation Directory

```bash
cd proj/debunk/toxic_sociopathy
mkdir -p evaluations/{intro,lie01,lie02,lie03,lie04,lie05}
```

### 3. Read the Documentation

**Required reading before starting:**
1. `evaluation_schema.md` - Understand the evaluation framework (30 min read)
2. `quick_reference.md` - Keep this open while evaluating (reference)
3. `evaluation_example.yaml` - See a complete example (10 min read)

### 4. Start Evaluating

**Phase 1: Triage (recommended start)**
- Go through all citations
- Mark accessibility status
- Flag obvious problems
- Estimated time: ~15 hours for 182 citations

**Phase 2: Quick Evaluation**
- Evaluate accessible sources
- Basic credibility + representation check
- Estimated time: ~30 hours

**Phase 3: Deep Dive**
- Detailed analysis of key citations
- Full verification and cross-referencing
- Estimated time: ~25-30 hours per 30 citations

---

## Files Overview

### Documentation Files

#### `evaluation_schema.md`
**Comprehensive evaluation methodology**
- 6 dimensions of evaluation
- Detailed category definitions
- Scoring system (optional)
- Common pitfalls to avoid
- ~8,000 words

**Use for**: Understanding the framework, reference during evaluation

#### `quick_reference.md`
**Cheat sheet for evaluators**
- Quick decision trees
- Common red flags
- Time budgets
- Keyboard shortcuts
- Pattern analysis checklist

**Use for**: Keep open while evaluating, quick lookups

#### `citations.md`
**Complete citation list with full text**
- All 222 citations as they appear in book
- Organized by section
- URLs included where readable
- Notes on unclear citations

**Use for**: Reference while evaluating, cross-checking book text

### Template Files

#### `evaluation_template.yaml`
**Blank template for evaluations**
- Copy this for each citation
- Fill in all fields
- Save in appropriate directory

**Use for**: Creating new evaluations

#### `evaluation_example.yaml`
**Fully worked example**
- Shows evaluation of NYC abortion statistics citation
- Demonstrates all fields filled in
- Example of credible source being misused

**Use for**: Learning how to fill out evaluations

---

## Evaluation Dimensions

Each citation is evaluated across **6 dimensions**:

### 1. Accessibility
Can we access the source?
- ACCESSIBLE, PAYWALLED, DEAD_LINK, etc.

### 2. Source Type
What kind of source is it?
- Peer-reviewed journal, mainstream news, advocacy org, etc.

### 3. Credibility
How reliable is the source? (1-5 scale)
- 5 = Highly credible (gov stats, peer-reviewed)
- 1 = Not credible (conspiracy sites, fabricated)

### 4. Representation Accuracy
Does the source say what the book claims?
- ACCURATE, CHERRY_PICKED, MISREPRESENTED, etc.

### 5. Bias Assessment
What biases affect the source?
- Political bias, funding, conflicts of interest

### 6. Meta-Assessment
Overall, how well does this support the book's argument?
- STRONG_SUPPORT, WEAK_SUPPORT, NO_SUPPORT

---

## Workflow Recommendations

### Option A: Section-by-Section (Recommended)
1. Pick one section (e.g., Lie #1 - Abortion)
2. Triage all citations in that section
3. Deep-evaluate the accessible ones
4. Write up findings for that section
5. Move to next section

**Pros**: Complete one section at a time, easier to track patterns
**Cons**: Takes longer to see overall picture

### Option B: Phased Approach
1. Triage ALL citations (accessibility only)
2. Quick-evaluate ALL accessible citations
3. Deep-dive on top 30 most important
4. Write up overall findings

**Pros**: Get overall picture faster, can identify patterns early
**Cons**: Context-switching between sections

### Option C: Random Sampling
1. Randomly select 30 citations across all sections
2. Deep-evaluate those 30
3. Use findings to decide if more sampling needed

**Pros**: Fastest way to get representative sample
**Cons**: Might miss section-specific patterns

---

## Key Statistics

### Citations by Section
- Introduction: 4 citations
- Lie #1 (Abortion): 36 citations
- Lie #2 (Trans): 58 citations
- Lie #3 (Love): 17 citations
- Lie #4 (Immigration): 73 citations
- Lie #5 (Social Justice): 34 citations
- **Total: 222 citations**

### Extracted URLs
- **~182 valid URLs** in `citations.yaml`
- **~45-50 URLs** commented out (need manual review)
  - Academic papers without direct URLs (~20-25)
  - Unclear/incomplete URLs (~15-20)
  - Book references (~5)
  - Suspicious/incorrect URLs (~2-3)

### Estimated Accessibility
Based on source types:
- ~40-50 citations: Fully accessible (free news, government reports)
- ~60-70 citations: Paywalled (NYT, WaPo, academic journals)
- ~30-40 citations: May be dead/removed (older social media, moved content)
- ~40-50 citations: No URL or unclear (books, unclear references)

**Realistically accessible**: ~100-120 citations out of 222

---

## Quality Metrics to Track

As you evaluate, track these metrics:

### Accessibility Metrics
- % fully accessible
- % paywalled (with/without preview)
- % dead links
- % no URL provided

### Source Type Distribution
- % academic/peer-reviewed
- % mainstream news
- % advocacy organizations
- % social media
- % government sources

### Credibility Metrics
- Average credibility score (1-5)
- % high credibility (4-5)
- % low credibility (1-2)
- Distribution by section

### Representation Metrics
- % accurately represented
- % cherry-picked
- % misrepresented
- % contradicts book's claim
- Common types of misrepresentation

### Red Flags
- Count of retracted sources
- Count of extremist sources
- Count of satire/fabricated sources
- Count of circular citations

### Overall Support
- % strong support
- % weak support
- % no support
- Average support by section

---

## Analysis Questions

After evaluation, you'll be able to answer:

### Source Quality
1. What percentage of citations are from credible sources?
2. Are there entire categories of evidence missing?
3. Does the book rely heavily on any particular low-quality source?
4. What's the average credibility score by section?

### Representation Quality
1. What percentage of citations are accurately represented?
2. What are the most common types of misrepresentation?
3. Are certain types of sources more likely to be misrepresented?
4. Are there patterns in cherry-picking?

### Bias Analysis
1. Does the book cite only from one political perspective?
2. What percentage of citations are from advocacy organizations?
3. Are mainstream/neutral sources represented fairly?
4. Do citations ignore contradictory evidence?

### Accessibility
1. Can readers actually verify the claims?
2. How many sources are behind paywalls?
3. How many sources are dead links or removed?
4. Is there transparency in sourcing?

### Meta Analysis
1. Do the sources actually support the book's main thesis?
2. Are the strongest claims backed by the weakest sources?
3. Is there a pattern of citation quality by topic?
4. What percentage of the book's argument rests on shaky sourcing?

---

## Output Recommendations

### For Each Section
Create a summary document:
- Total citations evaluated
- Accessibility breakdown
- Average credibility score
- Most common misrepresentation types
- Notable red flags
- Overall assessment

### Overall Report
Combine all sections:
- Executive summary
- Methodology
- Quantitative findings (charts/graphs)
- Qualitative findings (patterns, examples)
- Notable examples (best and worst citations)
- Conclusions
- Appendix (all individual evaluations)

### Presentation Formats
- **Academic**: Detailed methodology, statistical analysis
- **Blog post**: Highlight worst offenders with clear examples
- **Infographic**: Visual representation of statistics
- **Video**: Walk through specific examples
- **Social media**: Thread with key findings + examples

---

## Tools & Resources

### Credibility Assessment
- [Ad Fontes Media Bias Chart](https://adfontesmedia.com/)
- [AllSides Media Bias Ratings](https://www.allsides.com/media-bias)
- [Media Bias/Fact Check](https://mediabiasfactcheck.com/)
- [Retraction Watch](https://retractionwatch.com/)

### Finding Sources
- [Internet Archive Wayback Machine](https://archive.org/web/)
- [Google Scholar](https://scholar.google.com/)
- [Unpaywall](https://unpaywall.org/)
- [Sci-Hub](https://sci-hub.se/) (legal gray area)

### Fact Checking
- [Snopes](https://www.snopes.com/)
- [PolitiFact](https://www.politifact.com/)
- [FactCheck.org](https://www.factcheck.org/)
- [Climate Feedback](https://climatefeedback.org/)
- [Health Feedback](https://healthfeedback.org/)

### Journal Quality
- [Beall's List](https://beallslist.net/) (predatory journals)
- [DOAJ](https://doaj.org/) (legitimate open access)
- [Journal Impact Factors](https://www.scimagojr.com/)

---

## Best Practices

### Do
✅ Document everything
✅ Copy exact quotes from sources
✅ Note page numbers/timestamps
✅ Check for updates/retractions
✅ Look for contradictory evidence
✅ Be fair to sources you disagree with
✅ Take breaks to avoid bias creep
✅ Track time spent per citation
✅ Flag items for follow-up
✅ Stay organized (file naming, directories)

### Don't
❌ Rush through evaluations
❌ Skip inaccessible sources (mark them)
❌ Assume bias = unreliable
❌ Assume credentials = correct
❌ Evaluate when angry/emotional
❌ Cherry-pick your own evidence
❌ Ignore your own biases
❌ Make claims you can't back up
❌ Forget to save your work
❌ Work for too long without breaks

---

## Citation Naming Convention

Use consistent IDs for tracking:

**Format**: `{section}_{number}`

**Examples**:
- `intro_001` - Introduction, citation #1
- `lie01_019` - Lie #1, citation #19
- `lie02_043` - Lie #2, citation #43
- `lie04_067` - Lie #4, citation #67

**File naming**:
- `evaluations/lie01/lie01_019.yaml`
- `evaluations/lie02/lie02_043.yaml`

---

## Progress Tracking

Create a simple tracking file: `progress.md`

```markdown
# Evaluation Progress

## Phase 1: Triage (182 accessible citations)
- [ ] Introduction (4 citations)
- [ ] Lie #1 (31 citations)
- [ ] Lie #2 (37 citations)
- [ ] Lie #3 (17 citations)
- [ ] Lie #4 (64 citations)
- [ ] Lie #5 (28 citations)

## Phase 2: Quick Evaluation
- [ ] Introduction (X accessible)
- [ ] Lie #1 (X accessible)
...

## Phase 3: Deep Dive (30 key citations)
- [ ] Top priority citations identified
- [ ] Deep evaluations completed
...

## Metrics (update as you go)
- Citations evaluated: 0/182
- Average credibility: N/A
- Accurately represented: N/A
- Red flags found: 0
- Time spent: 0 hours
```

---

## Time Estimates

### Conservative Estimate
- **Triage**: 15 hours (5 min × 182)
- **Quick eval**: 30 hours (15 min × 120)
- **Deep dive**: 30 hours (60 min × 30)
- **Analysis**: 10 hours
- **Write-up**: 10 hours
- **Total**: ~95 hours

### Realistic Estimate (with learning curve)
- **Setup**: 5 hours
- **Triage**: 20 hours
- **Quick eval**: 40 hours
- **Deep dive**: 40 hours
- **Analysis**: 15 hours
- **Write-up**: 15 hours
- **Total**: ~135 hours

**Pace recommendations**:
- 5 hours/week: ~6 months
- 10 hours/week: ~3 months
- 20 hours/week: ~6 weeks

---

## Getting Started Checklist

Before you begin:
- [ ] Read `evaluation_schema.md` (understand framework)
- [ ] Read `evaluation_example.yaml` (see example)
- [ ] Bookmark `quick_reference.md` (for reference)
- [ ] Download citations with labrador tool
- [ ] Create evaluation directory structure
- [ ] Copy `evaluation_template.yaml` to use
- [ ] Set up progress tracking
- [ ] Schedule dedicated evaluation time
- [ ] Decide on workflow (section-by-section vs phased)
- [ ] Pick 3-5 citations to practice with

---

## Support & Questions

If you need help:
1. Re-read the relevant section of `evaluation_schema.md`
2. Check `quick_reference.md` for quick answers
3. Look at `evaluation_example.yaml` for guidance
4. Take a break and come back fresh
5. Ask for clarification on ambiguous cases

---

## Version History

- **2026-06-17**: Initial creation
  - Evaluation schema defined
  - Template and example created
  - Quick reference guide added
  - All 222 citations extracted and documented

---

## Next Steps

1. ✅ Extract citations from photos (DONE)
2. ✅ Create evaluation framework (DONE)
3. ⏭️ Download citations with labrador
4. ⏭️ Begin triage phase
5. ⏭️ Conduct evaluations
6. ⏭️ Analyze patterns
7. ⏭️ Write up findings
8. ⏭️ Publish results

Good luck with your analysis! 🔍
