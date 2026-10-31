# Curriculum gap analysis

Audit date: 2026-10-30. Reviewed every module in `curriculum/helicopter.json` against the project brief. This is a content inventory and instructional-design assessment, not independent validation of engineering accuracy or bibliography.

## Inventory

| Stage | Modules | Estimated hours |
| --- | ---: | ---: |
| S1 Foundations | 5 | 350 |
| S2 Mechanical Engineering | 6 | 500 |
| S3 CAD and Manufacturing | 5 | 300 |
| S4 Rotorcraft Aerodynamics | 6 | 500 |
| S5 Helicopter Mechanical Systems | 6 | 500 |
| S6 Dynamics and Avionics | 5 | 400 |
| S7 Safety and Certification | 3 | 250 |
| S8 Helicopter Capstone | 6 | 400 |
| **Total** | **42** | **3,200** |

The stage hour totals equal their module sums. All IDs are unique; all prerequisite references exist and form an acyclic graph. The existing dependency graph is one 42-module chain: the first module has no prerequisite and every subsequent module depends on the immediately preceding one. It is mechanically valid but not a well-developed subject dependency model. For example, every CAD module waits for all six mechanical modules; basic flight controls wait for FEA; thermodynamics waits for composites. Review educational dependencies before turning these into hard study locks.

## Requirements coverage

Counts below describe explicit structured curriculum data, not inferred competency coverage.

| Required element | Supplied ZIP | Verified baseline / remaining gap |
| --- | --- | --- |
| Unique ID, stage, module title, estimated hours | 42/42 | Retained; validated. Subject classification implicit in titles/stages. Hours have not been validated as realistic. |
| Prerequisites | 42/42 arrays (41 nonempty) | Retained, validated and enforced on pass. Subject DAG/diagnostic placement still needed. |
| Learning objectives | 0/42 dedicated objective lists | `criteria` contains outcome-like prose, not complete objectives. Define measurable objectives separately. |
| Required theory | 0/42 structured theory lists | Topics are implied by titles/resources. Add concepts, assumptions, notation and prior knowledge. |
| Required engineering calculations | 0/42 dedicated calculation lists | Many criteria name calculations; no given datasets, expected units, tolerances or reference answers. |
| Primary textbook resources | 42/42 prose resource strings | Many are abbreviated author names or mixed textbook/course guidance; not 42 fully specified bibliographies. |
| Verified exact edition/chapter/section assignments | 0/42 verification records | All 42 flagged `requires_verification`. A few inherited chapter claims removed. No precise assignments added. |
| Additional lectures/tutorials | Some prose references; no structured lecture registry | Add verified links and exact lecture/topic scope; availability/access not checked in this baseline. |
| Practical exercises | 0/42 exercise sets | Some criteria suggest CAD/code/report deliverables; no step-by-step exercises or datasets. |
| Engineering assignments | 0/42 complete assignment briefs | Define inputs, constraints, deliverables, assumptions and review questions. |
| Assessment methods | 42/42 identical generic paragraph | All request calculations/practical work and ≥80. No module-specific rubric, question bank, solutions or critical-error definition. |
| Explicit passing criteria | 42/42 competency paragraphs | Useful starting points; measurable acceptance tests and mandatory item lists still needed. Baseline enforces overall score, evidence, practical confirmation and prerequisites. |
| References for verification | 0/42 source/provenance records | Add publisher/author/course URL, edition, reading location, date checked and verification note. |
| Completion status | Static `not_started` plus runtime progress model | Removed static learner status from content. Five runtime states, score/evidence/practical confirmation, timestamps and sessions now work. |
| Submitted work / portfolio | Notes only; no artifact model | Baseline adds plain-text evidence references. Typed artifacts, revisions, design decisions and requirement links remain planned. |
| Assessment history and review | Latest status/score only | Baseline remains latest-only; no attempt history, reviewer identity, rubric version or submission timestamp. |

All modules still need structured objectives, theory/calculation lists, verified resource metadata, exercises, assignment briefs and specific rubrics. The module register below highlights each module's **additional priority**; it does not repeat those universal gaps.

## Module register: specific next content work

| ID | Current module | Priority gap / next acceptance evidence |
| --- | --- | --- |
| S1M1 | Calculus and differential equations | Prerequisite diagnostic for algebra/trigonometry/units; scaffold derivatives, integrals, multivariable work and ODEs instead of assuming all can be learned in 55 hours. Analytical-vs-RK4 comparison with error bounds. |
| S1M2 | Linear algebra and coordinate frames | Define frame/rotation conventions and matrix dimensions; transformation test cases, orthogonality and eigenproblem reference values. |
| S1M3 | Statics and free-body diagrams | Provide rotor-support geometry/load cases; distinguish forces/couples and require independent equilibrium residual checks. |
| S1M4 | Rigid-body dynamics | Specify inertia/gyroscopic sign conventions and worked datasets; conservation and dimensional cross-checks. |
| S1M5 | Numerical methods and uncertainty | Separate numerical error from measurement uncertainty; convergence tolerances, sensitivity cases and a statistics/probability foundation. |
| S2M1 | Strength of materials | Define beam/shaft boundary conditions, allowables and load cases; compare calculations against analytical references. |
| S2M2 | Machine elements and fatigue | Two-stage gearbox design is too broad without staged fatigue/shaft/fastener/bearing problems; add material data provenance and load spectra. |
| S2M3 | Gears, lubrication and tribology | Define gear geometry, stress model applicability, oil/thermal assumptions and example rating checks; coordinate overlap with S2M2/S5M3. |
| S2M4 | Materials, fracture and composites | Separate material selection, fracture and laminate calculations; provide validated properties, layup conventions and failure limitations. |
| S2M5 | Thermodynamics and heat transfer | Provide cycle/control-volume diagrams, property data and gearbox cooling cases; close mass/energy balances explicitly. |
| S2M6 | Fluid mechanics and vibration | Split unrelated foundations or sequence them explicitly; hydraulic and SDOF datasets with loss/damping model limits. |
| S3M1 | SolidWorks parts and design intent | Define the ten parts, design-intent changes, constraint checks and required native/drawing exports; distinguish tutorial reproduction from independent modeling. |
| S3M2 | Assemblies and motion | Supply swashplate kinematic brief, intended DOF, mate strategy and collective/cyclic test cases before requiring a full assembly. |
| S3M3 | Engineering drawings and GD&T | Verify standard edition; provide datum/function examples, fits, quantitative tolerance stack and inspection criteria. |
| S3M4 | Manufacturing and inspection | Define five representative helicopter parts, process constraints, inspection uncertainty and report/drawing deliverables. |
| S3M5 | Finite element analysis | Add stiffness derivation and cantilever reference problem, boundary-condition justification, element/mesh convergence tolerances and result limitations. |
| S4M1 | Helicopter systems and controls | Move an introductory orientation earlier; verify handbook revision/topics and supply a weight-and-balance dataset. |
| S4M2 | Airfoil aerodynamics | Provide verified airfoil data with Reynolds/Mach ranges, interpolation limits and stall/compressibility explanation checks. |
| S4M3 | Momentum theory and hover performance | Pilot detailed module: induced-velocity/power derivation, numerical thrust/disk-area/hover-power case, Python implementation, independent calculation and ideal-vs-actual losses. Verify Leishman edition/readings first. |
| S4M4 | Blade element and BEMT | Specify airfoil/polar inputs, solver equations, nondimensional coefficients, convergence thresholds and analytical limiting-case checks. |
| S4M5 | Forward flight and trim | Define velocity/azimuth/frame conventions, simplifications, trim unknowns and residual acceptance; distinguish model limitations from verified performance. |
| S4M6 | Unsteady aerodynamics and autorotation | Separate wake/inflow and autorotation concepts; add energy-flow sign checks and a bounded dynamic-response reference case. |
| S5M1 | Blade structures and aeroelasticity | Provide blade structural/mass model, centrifugal/bending cases and modal benchmark; define what a simplified flutter model cannot establish. |
| S5M2 | Rotor hub and swashplate | Define articulated-system kinematics and motion-envelope deliverables; connect CAD revisions to loads and interface assumptions. |
| S5M3 | Main gearbox and rotor drive | Integrate S2 machine-element work with explicit gearbox load/thermal/lubrication cases and failure-path review rubric. |
| S5M4 | Tail rotor and anti-torque drive | Define moment arms, main-rotor reaction torque and control assumptions; include parameter sensitivities and power-budget consistency. |
| S5M5 | Turboshaft and engine integration | Distinguish ideal cycle from actual engine maps, installation derating and power limits; cite data origins. |
| S5M6 | Fuselage, landing gear and hydraulics | Split airframe, landing energy and actuator work into traceable briefs; define loads/boundaries and separate preliminary estimates from detailed substantiation. |
| S6M1 | Rigid body flight dynamics and stability | Define six-DOF conventions, trim/linearization procedure, numerical reference cases and eigenvalue interpretation. |
| S6M2 | Feedback and digital control | Add sampling/discretization assumptions, actuator limits, anti-windup cases and quantified stability/response criteria. |
| S6M3 | Sensors and Kalman filtering | Supply deterministic simulated sensor datasets/noise assumptions; require tuning rationale, drift/error metrics and reproducibility. |
| S6M4 | Embedded real-time avionics | Define task rates/deadlines, scheduling model, fault assumptions, watchdog behavior and measured timing evidence. |
| S6M5 | Integrated simulation | Provide model interfaces, trim/reference cases, fault matrix and reproducibility expectations; avoid treating simulation as aircraft validation. |
| S7M1 | Systems engineering and traceability | Define requirement-quality rules and example subsystem interface/verification matrix, including assumptions and test methods. |
| S7M2 | Reliability and safety assessment | Verify relevant source/standard editions; supply bounded FHA/FMEA/fault-tree examples, consistency checks and critical-error rubric. |
| S7M3 | Regulatory framework and compliance | Verify current jurisdiction/applicability/revisions before assigning clauses; distinguish classroom compliance mapping from certification authority approval. |
| S8M1 | Mission requirements and architecture | The two-seat 700 kg brief is illustrative, not a validated design. Define mission assumptions, budgets, trade studies and review gates. |
| S8M2 | Rotor preliminary sizing | Require a reproducible parameter set, uncertainty/sensitivity record and links back to checked BEMT/momentum work. |
| S8M3 | Powertrain and rotor system CAD | Require revisioned interfaces, loads, BOM, kinematic checks and traceability to calculations; no unlimited binary storage. |
| S8M4 | Airframe and integration | Define equipment envelopes, CG/loading cases, interface drawings, tolerances and integration open issues. |
| S8M5 | Controls and system simulation | Link model/control revisions to scenarios and acceptance metrics; track failed cases and unresolved modeling limits. |
| S8M6 | Independent design review | Define reviewer independence, dossier contents, review questions and resolution records; preserve open safety issues rather than converting a score to airworthiness. |

## Cross-cutting educational gaps

1. **Entry foundations:** algebra, trigonometry, vectors, units/dimensional analysis and basic physics are assumed rather than diagnosed/taught. The learner's CS/embedded experience can support numerical/avionics work, but should not waive mechanics or mathematics assessments without evidence.
2. **Experimental engineering:** measurement, instrumentation, uncertainty propagation, data reduction, verification-vs-validation, bench-test planning and test reporting need explicit learning outcomes across stages.
3. **Integration:** mass/CG, load envelopes, power/thermal budgets, subsystem interfaces and configuration management need continuity across assignments and capstone versions.
4. **Educational scaffolding:** a large module/topic title is not a teachable sequence. Add worked examples, practice with feedback, independent assignments and assessment rubrics in small verified batches.
5. **Evidence and sources:** the baseline records references but cannot inspect them. A reviewed artifact registry and precise source provenance are necessary before calling this a comprehensive engineering education platform.

No textbook chapters, engineering test results, assessment grades or source-verification outcomes were invented in this audit. Verification of educational resources is the next content milestone, not a completed baseline feature.
