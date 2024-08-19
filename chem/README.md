Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/chem)

Package `chem` provides basic chemistry simulation algorithms, including:

* `React` -- chemical reaction where 2 components bind into a compound, characterized by forward and backward rate constants, Kf, Kb.

* `Enz` -- enzyme-catalyzed reaction based on the Michaelis-Menten kinetics that transforms S = substrate into P product via SE bound C complex.

* `Buffer` -- provides a soft buffering driving deltas relative to a target N which can be set by concentration and volume.

* `Integrate` -- performs basic forward Euler integration, using `IntegrateDt` rate constant.

* `CoFmN` and `CoToN` convert concentration to / from numbers of molecules given a volume (just multiplication and division, but useful for documenting the purpose).

In general, all of this code just computes deltas (discrete derivatives) for moving numbers of molecules around -- complex systems of reactions can be constructed and all the different deltas summed up, and applied in a step-wise fashion.  At bottom, it is really very simple, involving massive simplifications that nevertheless seem sufficient to capture the relevant phenomena.  In effect, the all-important rate constants absorb and compensate for all of those simplifications.

# Kinetikit

This code is based on [Kinetikit](https://www.ncbs.res.in/faculty/bhalla-kinetikit) by Upinder S. Bhalla and implemented in the [Genesis](http://genesis-sim.org) simulation tool.  See:

* Bhalla, U. S., & Iyengar, R. (1999). Emergent Properties of Networks of Biological Signaling Pathways. Science. https://doi.org/10.1126/science.283.5400.381

See the [axon urakubo](https://github.com/emer/axon/tree/main/examples/urakubo) model of LTP / LTD for a re-implementation of the Urakubo et al, 2008 model using this code.

* Urakubo, H., Honda, M., Froemke, R. C., & Kuroda, S. (2008). Requirement of an allosteric kinetics of NMDA receptors for spike timing-dependent plasticity. The Journal of Neuroscience, 28(13), 3310–3323. http://www.ncbi.nlm.nih.gov/pubmed/18367598

Here's another paper that Urakubo builds upon:

* Dupont, G., Houart, G., & De Koninck, P. (2003). Sensitivity of CaM kinase II to the frequency of Ca2+ oscillations: A simple model. Cell Calcium, 34(6), 485–497. https://doi.org/10.1016/S0143-4160(03)00152-0


# Reactions

`React` models a basic chemical reaction:

```
      Kf
A + B --> AB
     <-- Kb
```

where Kf is the forward and Kb is the backward time constant.  The source Kf and Kb constants are in terms of concentrations μM-1 and sec-1 but calculations take place using N's, and the forward direction has two factors while reverse only has one, so a corrective volume factor needs to be divided out to set the actual forward factor.

# Enzymes

`Enz` models an enzyme-catalyzed reaction based on the Michaelis-Menten kinetics that transforms S = substrate into P product via SE bound C complex:

```
      K1        K3
S + E --> C(SE) ---> P + E
     <-- K2
```

S = substrate, E = enzyme, C = SE complex, P = product.  The source K constants are in terms of concentrations μM-1 and sec-1 but calculations take place using N's, and the forward direction has two factors while reverse only has one, so a corrective volume factor needs to be divided out to set the actual forward factor.

