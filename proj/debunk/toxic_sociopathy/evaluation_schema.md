# Source Evaluation Schema

This document outlines the categorization system for evaluating the factuality and quality of citations.

---

## Overview

Each source should be evaluated across multiple dimensions. The goal is to determine:
1. **Can we access the source?** (Accessibility)
2. **What type of source is it?** (Source Type)
3. **How credible is the source?** (Credibility)
4. **Does it say what the book claims?** (Representation Accuracy)
5. **What biases does it have?** (Bias Assessment)

---

## Dimension 1: Accessibility

**Question**: Can we actually access and read the source?

### Categories

- **ACCESSIBLE**: Source fully accessible, content retrieved successfully
- **PAYWALLED**: Source exists but behind paywall (partial access via preview/abstract)
- **PAYWALLED_NO_ACCESS**: Source completely paywalled, no preview available
- **DEAD_LINK**: URL returns 404 or similar error
- **REDIRECT**: URL redirects to different content (note where it redirects)
- **REMOVED**: Content was removed by publisher/platform (check Internet Archive)
- **NO_URL**: Citation has no URL (book, report, etc. - requires library access)
- **REQUIRES_AUTH**: Requires login/account (Twitter, Facebook, academic databases)

### Notes
- For PAYWALLED sources, note if you have institutional access
- For DEAD_LINK/REMOVED, check Internet Archive Wayback Machine
- For NO_URL, attempt to locate via library, Google Scholar, or purchase

---

## Dimension 2: Source Type

**Question**: What kind of source is this?

### Categories

#### Academic/Research
- **PEER_REVIEWED_JOURNAL**: Published in peer-reviewed academic journal
- **ACADEMIC_PREPRINT**: Academic paper not yet peer-reviewed (arXiv, bioRxiv, SSRN)
- **ACADEMIC_BOOK**: Scholarly book from academic press
- **DISSERTATION**: PhD dissertation or thesis
- **ACADEMIC_REPORT**: Research report from university or research institution

#### News/Media
- **MAINSTREAM_NEWS**: Established newspaper or news organization (NYT, WaPo, Reuters, AP)
- **CABLE_NEWS**: Cable news network (CNN, Fox News, MSNBC)
- **ALTERNATIVE_NEWS**: Alternative or independent news outlets
- **OPINION_PIECE**: Editorial or opinion column (clearly marked as opinion)
- **BLOG**: Personal or organizational blog

#### Government/Official
- **GOV_REPORT**: Official government report or statistics
- **GOV_DATABASE**: Government database or data portal (CDC, FBI, Census)
- **COURT_DOCUMENT**: Court filing, decision, or transcript
- **LEGISLATIVE**: Congressional testimony, bill text, etc.

#### Advocacy/Think Tank
- **ADVOCACY_ORG**: Advocacy organization with clear political stance
- **THINK_TANK**: Policy research organization (note political leaning)
- **SPECIAL_INTEREST**: Special interest group or lobbying organization

#### Social Media/Primary
- **SOCIAL_MEDIA**: Twitter, Facebook, Instagram post
- **PRESS_RELEASE**: Official press release
- **PRIMARY_DOCUMENT**: Original historical document, letter, etc.
- **VIDEO**: YouTube or other video content
- **PODCAST**: Audio podcast episode

#### Other
- **ENCYCLOPEDIA**: Wikipedia, Britannica, specialized encyclopedias
- **FACT_CHECK**: Snopes, PolitiFact, FactCheck.org
- **BOOK**: Non-academic trade book
- **MAGAZINE**: Magazine article (Time, Atlantic, etc.)

### Notes
- A source can have multiple types (e.g., OPINION_PIECE + MAINSTREAM_NEWS)
- Note the specific outlet/publisher for context

---

## Dimension 3: Source Credibility

**Question**: How reliable and credible is this source?

### Rating Scale (1-5)

#### 5 - Highly Credible
- Peer-reviewed academic journals with rigorous methodology
- Official government statistics from reputable agencies (CDC, Census Bureau, FBI)
- Established fact-checking organizations
- Primary source documents (court records, official transcripts)
- Multiple independent corroborating sources

**Indicators**: Rigorous methodology, transparent data, expert consensus, independent verification

#### 4 - Generally Credible
- Mainstream news organizations with fact-checking (NYT, WaPo, Reuters, AP)
- Reputable think tanks with transparent methodology
- Academic books from university presses
- Government reports from credible agencies
- Well-documented investigative journalism

**Indicators**: Editorial standards, corrections policy, named sources, verifiable claims

#### 3 - Mixed Credibility
- News sources with known political bias but factual reporting
- Advocacy organizations citing verifiable data
- Opinion pieces by subject matter experts
- Secondary reporting without original sources
- Sources with occasional factual errors but general reliability

**Indicators**: Some bias present, relies on other sources, selective emphasis, mostly verifiable

#### 2 - Low Credibility
- Highly partisan sources with poor fact-checking
- Opinion pieces presented as news
- Sources with history of retractions/corrections
- Advocacy organizations with minimal sourcing
- Blogs without expertise or verification
- Social media posts without verification

**Indicators**: Heavy bias, anonymous sources, unverifiable claims, sensationalism, poor sourcing

#### 1 - Not Credible
- Sources known for spreading misinformation
- Conspiracy theory websites
- Fabricated or satirical content presented as news
- Sources with proven track record of false claims
- Completely unverified social media rumors

**Indicators**: No editorial standards, false claims, conspiracy theories, no accountability

### Special Cases

- **RETRACTED**: Academic paper that has been retracted (automatically = 1)
- **CORRECTED**: Source issued significant correction (note impact on credibility)
- **DISPUTED**: Claims in source are actively disputed by experts
- **OUTDATED**: Source predates current scientific consensus

### Notes
- Credibility is about the source's general reliability, not this specific claim
- A credible source can still be misrepresented
- Note any conflicts of interest (funding, author affiliations)

---

## Dimension 4: Representation Accuracy

**Question**: Does the source actually say what the book claims it says?

### Categories

#### ACCURATE
Source directly supports the book's claim with proper context.

**Example**: Book claims "Study found X", source abstract clearly states "We found X"

#### ACCURATE_BUT_CHERRY_PICKED
Source technically supports the claim but omits important context, limitations, or contradictory findings in the same source.

**Example**: 
- Book cites one statistic from a study
- Source also contains qualifications, limitations, or contradictory data
- Book ignores the "however" or "but" in the original

#### MISLEADING_CONTEXT
Source is real but the book removes critical context that changes the meaning.

**Example**:
- Original: "Some studies suggest X, however the evidence is limited and contradicted by Y"
- Book: "Studies show X"

#### MISLEADING_OUTDATED
Source was accurate when published but has been superseded by more recent evidence. Book presents it as current.

**Example**: 
- Book cites 2008 study
- Multiple 2020+ studies contradict the 2008 findings
- Book doesn't mention newer research

#### PARTIAL_TRUTH
Source contains elements of the claim but also significant nuance or contradiction that the book ignores.

**Example**:
- Source discusses complex issue with multiple perspectives
- Book extracts only one perspective as definitive

#### MISREPRESENTED
Source says something different from what the book claims, either through:
- Misquoting
- Taking quote out of context
- Misinterpreting data/findings
- Confusing correlation with causation

**Example**:
- Source: "No evidence was found for X"
- Book: "Source confirms X is not a concern"

#### CONTRADICTS
Source actually contradicts or disputes the book's claim.

**Example**:
- Book uses source to support claim X
- Source explicitly argues against X

#### INACCESSIBLE
Cannot verify because source is not accessible (see Accessibility dimension).

#### NOT_RELEVANT
Source is real and accessible but doesn't actually relate to the claim being made.

**Example**:
- Book makes claim about policy X
- Source discusses different policy Y

### Sub-Classifications

For any category except ACCURATE, note:

1. **Severity**: Minor, Moderate, Severe
   - **Minor**: Small omission that doesn't fundamentally change meaning
   - **Moderate**: Significant context missing but core fact remains
   - **Severe**: Fundamentally misrepresents source or reverses meaning

2. **Type of Misrepresentation**:
   - **QUOTE_OUT_OF_CONTEXT**: Direct quote used without surrounding context
   - **DATA_MISINTERPRETED**: Numbers/statistics misread or misapplied
   - **CAUSATION_CONFUSED**: Correlation presented as causation
   - **AUTHOR_INTENT_REVERSED**: Source making opposite argument
   - **SAMPLING_IGNORED**: Ignores that source studied specific population
   - **LIMITATIONS_IGNORED**: Ignores caveats/limitations in source
   - **RETRACTED_IGNORED**: Uses retracted or discredited research

### Notes
- Include specific quotes from both book and source
- Note page numbers/timestamps for precise reference
- For nuanced cases, quote the full context from source

---

## Dimension 5: Bias Assessment

**Question**: What potential biases affect this source?

### Political Bias

Use a scale or established ratings:
- **LEFT**: Ad Fontes Media or AllSides ratings
- **CENTER_LEFT**
- **CENTER**
- **CENTER_RIGHT**
- **RIGHT**
- **EXTREME_LEFT** / **EXTREME_RIGHT**: Highly partisan, potentially extremist

### Funding/Conflict of Interest

- **CORPORATE_FUNDED**: Source funded by corporations with stake in outcome
- **POLITICALLY_FUNDED**: Funded by political organizations/PACs
- **ADVOCACY_FUNDED**: Funded by advocacy groups
- **GOVERNMENT_FUNDED**: Government funding (note which government/agency)
- **INDEPENDENT**: No apparent conflicts
- **UNKNOWN_FUNDING**: Funding sources unclear

### Author/Institutional Bias

- **SUBJECT_MATTER_EXPERT**: Author has relevant credentials
- **LACKS_EXPERTISE**: Author lacks credentials in field
- **ACTIVIST**: Author is activist in this area (note: not inherently disqualifying)
- **INSTITUTIONAL_BIAS**: Author's institution has known stance

### Methodological Concerns

For research/studies only:
- **SMALL_SAMPLE**: Sample size too small for conclusions
- **POOR_METHODOLOGY**: Methodological flaws noted by reviewers
- **NON_REPRESENTATIVE**: Sample not representative of claimed population
- **CONFLICT_REVIEWED**: Peer reviewers noted conflicts/concerns
- **PREREGISTERED**: Study was pre-registered (positive indicator)
- **REPLICATED**: Findings have been replicated (positive indicator)
- **FAILED_REPLICATION**: Study failed to replicate

### Notes
- Bias doesn't automatically mean unreliable
- Note whether bias is transparent vs hidden
- Multiple independent biased sources can still establish facts

---

## Dimension 6: Meta-Assessment

**Question**: Overall, how does this citation serve the book's argument?

### Categories

#### STRONG_SUPPORT
- Highly credible source
- Accurately represented
- Directly supports claim
- Minimal bias or transparent bias
- Current/up-to-date

#### WEAK_SUPPORT
- Source technically supports claim but:
  - Low credibility, OR
  - Important context missing, OR
  - Outdated, OR
  - Significant methodological concerns

#### NO_SUPPORT
- Source doesn't actually support the claim
- Misrepresented or taken out of context
- Contradicts the claim
- Not relevant to the claim

#### INDETERMINATE
- Cannot access source to verify
- Source is ambiguous
- Need additional sources to contextualize

### Red Flags

Mark if any of these apply:
- ⚠️ **RETRACTED**: Research has been retracted
- ⚠️ **DISCREDITED**: Source has been widely discredited
- ⚠️ **EXTREMIST**: Source from extremist/hate group (SPLC designation, etc.)
- ⚠️ **SATIRE**: Satirical content presented as fact
- ⚠️ **MISATTRIBUTED**: Quote/claim attributed to wrong person/source
- ⚠️ **FABRICATED**: Source appears to be fabricated
- ⚠️ **CIRCULAR**: Source cites the book's author or related work (circular reasoning)

---

## Data Structure Recommendation

For tracking evaluations, use this YAML structure:

```yaml
citation_id: "lie01_005"
section: "Lie #1: Abortion Is Health Care"
citation_number: 5
url: "https://www.nbcnews.com/id/wbna4572168"
full_citation: "Doctor Investigated in 'Really Botched Abortion,' NBC News, February 5, 2009"

accessibility:
  status: ACCESSIBLE
  notes: "Successfully downloaded"
  archive_url: "" # If using Wayback Machine

source_type:
  primary: MAINSTREAM_NEWS
  secondary: []
  outlet: "NBC News"

credibility:
  rating: 4
  reasoning: "Mainstream news source with editorial standards"
  conflicts: []
  
representation:
  accuracy: ACCURATE
  severity: null
  misrep_type: []
  book_claim: "Doctor investigated for botched abortion resulting in live birth"
  source_quote: "The doctor is being investigated for what appears to be a severely botched abortion..."
  notes: "Claim accurately represents news story"

bias:
  political: CENTER
  funding: CORPORATE_FUNDED
  author_expertise: JOURNALIST
  methodological_concerns: []

meta:
  support: STRONG_SUPPORT
  red_flags: []
  overall_notes: "Credible news report accurately cited"
  
verification:
  verified_by: "Your Name"
  verification_date: "2026-06-17"
  requires_followup: false
  followup_notes: ""
```

---

## Practical Workflow

### Phase 1: Triage (Quick Pass)
For each citation:
1. Check accessibility (5 min)
2. Identify source type (2 min)
3. Flag obvious red flags (retracted, extremist, etc.)

**Output**: Prioritize which sources need deep investigation

### Phase 2: Credibility Assessment (Medium Pass)
For accessible, non-flagged sources:
1. Assess source credibility (10 min)
2. Identify obvious bias/conflicts
3. Quick-read to verify representation accuracy

**Output**: "Quick-check" rating for each source

### Phase 3: Deep Verification (Detailed Pass)
For sources the book relies on heavily:
1. Detailed read with highlighting
2. Compare specific claims to source
3. Research source credibility (retractions, author background)
4. Check for newer research/updates
5. Look for contradictory evidence

**Output**: Detailed report for key claims

### Phase 4: Pattern Analysis
Across all sources:
1. What % are accessible?
2. What % are accurately represented?
3. Are there patterns in credibility levels?
4. Which types of sources are over-represented?
5. Are there entire categories of evidence ignored?

**Output**: Meta-analysis of citation quality

---

## Scoring System (Optional)

For quantitative analysis, you could score each citation:

### Component Scores
- **Accessibility**: 0-2 points
  - 2 = ACCESSIBLE
  - 1 = PAYWALLED (partial access)
  - 0 = INACCESSIBLE
  
- **Credibility**: 1-5 points
  - Use the 1-5 scale above

- **Representation**: 0-3 points
  - 3 = ACCURATE
  - 2 = ACCURATE_BUT_CHERRY_PICKED
  - 1 = MISLEADING (any type)
  - 0 = MISREPRESENTED / CONTRADICTS

- **Red Flag Penalty**: -5 points
  - If any major red flag present

### Total Score
- **Maximum**: 10 points (highly credible, accurately cited)
- **Minimum**: -5 points (red flags present)

### Interpretation
- **8-10**: Strong citation
- **5-7**: Acceptable but flawed
- **2-4**: Weak citation
- **0-1**: Very poor citation
- **Negative**: Actively misleading

### Book-Level Score
- Average all citations
- Calculate by section
- Weight by importance (optional)

---

## Red Herrings to Avoid

1. **Ad Hominem Trap**: Bias doesn't make facts false
   - Even biased sources can report accurate facts
   - Focus on verifiability, not source politics

2. **Credentialism Trap**: Credentials aren't everything
   - An MD can be wrong about medicine
   - A non-expert can accurately report a study

3. **Recency Trap**: Old ≠ wrong
   - Historical sources may be perfectly valid for historical claims
   - Only flag as outdated if superseded by new evidence

4. **Quantity Trap**: 100 bad sources don't make a good argument
   - Quality matters more than quantity
   - Look for original research vs media reports of research

5. **Perfect Source Fallacy**: No source is perfect
   - Even peer-reviewed studies have limitations
   - Note flaws without dismissing entirely

---

## Useful Tools & Resources

### Credibility Assessment
- **Media Bias Chart**: Ad Fontes Media (bias + reliability)
- **AllSides**: Media bias ratings
- **MBFC**: Media Bias/Fact Check
- **Retraction Watch**: Tracking retracted papers
- **Google Scholar**: Citation counts, follow-up research
- **PubPeer**: Post-publication peer review

### Fact-Checking
- **Snopes**: General fact-checking
- **PolitiFact**: Political claims
- **FactCheck.org**: Political claims
- **Climate Feedback**: Climate science
- **Health Feedback**: Health/medical claims

### Access
- **Internet Archive**: Wayback Machine for dead links
- **Sci-Hub**: Academic papers (legal gray area)
- **unpaywall**: Legal open access papers
- **Google Scholar**: Often links to free PDFs
- **Library Access**: University or public library databases

### Analysis
- **CiteScore / Impact Factor**: Journal prestige
- **H-Index**: Author influence
- **CORE**: Journal ranking
- **Beall's List**: Predatory journals

---

## Example Evaluation

See `evaluation_example.yaml` for a fully worked example.
