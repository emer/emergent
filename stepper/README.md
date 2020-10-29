Visit the [GoDoc](https://godoc.org/github.com/emer/emergent/stepper) link for a detailed description of the `Stepper` API.

# Stepper

The `stepper` package provides a facility for pausing a simulation run at various time scales without having to
worry about saving and restoring simulation state.

## How to use the Stepper

1. Import the `stepper` package:

    ```go
   import "github.com/emer/emergent/stepper"

2. Define an enumerated type for whatever different types/granularities of steps you'd like. Note that the stepper
does not interpret these values, and only checks equality to see decide whether or not to stop at any given StepPoint.

   ```go
   type StepGrain int

   const (
	     Cycle StepGrain = iota
	     Quarter
	     SettleMinus
	     SettlePlus
	     AlphaCycle
	     SGTrial // Trial
	     TrialGroup
	     RunBlock
	     StepGrainN
   )
   //go:generate stringer -linecomment -type=StepGrain

   var KiT_StepGrain = kit.Enums.AddEnum(StepGrainN, kit.NotBitFlag, nil)

3. Add a few fields to your `Sim` struct:

   ```go
   struct Sim {
      Stepper                      *stepper.Stepper  `view:"-"`        
      StepsToRun                   int               `view:"-" desc:"number of steps to execute before stopping"`
      OrigSteps                    int               `view:"-" desc:"saved number of steps to execute before stopping"`
      StepGrain                    StepGrain         `view:"-" desc:"granularity for the Step command"`
      StopStepCondition            StopStepCond      `desc:"granularity for conditional stop"`
      StopConditionTrialNameString string            `desc:"if StopStepCond is TrialName or NotTrialName, this string is used for matching the current AlphaTrialName"`
      StopStepCounter              env.Ctr           `inactive:"+" view:"-" desc:"number of times we've hit whatever StopStepGrain is set to'"`
   }

4. Define a `stepper.PauseNotifyFn` callback:

   ```go
   // NotifyPause is called from within the Stepper, with the Stepper's lock held.
   // This function's code must not call any Stepper methods (e.g. Enter) that try to take the Stepper mutex lock.
   func (ss *Sim) NotifyPause() {
      if int(ss.StepGrain) != ss.Stepper.Grain() {
         ss.Stepper.StepGrain = int(ss.StepGrain)
      }
      if ss.StepsToRun != ss.OrigSteps { // User has changed the step count while running
        ss.Stepper.StepsPer = ss.StepsToRun
        ss.OrigSteps = ss.StepsToRun
      }
      ss.IsRunning = false
      ss.ToolBar.UpdateActions()
      ss.UpdateView()
      ss.Win.Viewport.SetNeedsFullRender()
   }

5. (__OPTIONAL__) Create a `stepper.StopCheckFn` callback:

    ```go
   // CheckStopCondition is called from within the Stepper, with the Stepper's lock held.
   // This function's code must not call any Stepper methods (e.g. Enter) that try to take the Stepper mutex lock.
    func (ss *Sim) CheckStopCondition(_ int) bool {
       ev := &ss.Env
       ret := false
       switch ss.StopStepCondition {
          case SSNone:
             return false
          case SSError:
             ret = ss.SumSSE > 0.0
          case SSCorrect:
             ret = ss.SumSSE == 0.0
          case SSTrialNameMatch:
             ret = strings.Contains(ev.AlphaTrialName, ss.StopConditionTrialNameString)
          case SSTrialNameNonmatch:
             ret = !strings.Contains(ev.AlphaTrialName, ss.StopConditionTrialNameString)
          default:
             ret = false
       }
       return ret
    }

6. Somewhere in your initialization code, create the actual `Stepper` and register your `stepper.PauseNotifyFn`
and (optionally) `stepper.StopCheckFn` functions:

   ```go
   ss.Stepper = stepper.New()
   ss.Stepper..StopCheckFn = CheckStopCondition
   ss.Stepper.PauseNotifyFn = NotifyPause

7. At appropriate points in your simulation code, insert `stepper.StepPoint` calls, e.g.:

   ```go
   func (ev *PVLVEnv) RunOneTrial(ss *Sim, curTrial *data.TrialInstance) {
      trialDone := false
      for !trialDone {
         ev.SetupOneAlphaTrial(curTrial, 0)
         ev.RunOneAlphaCycle(ss, curTrial)
         trialDone = ev.AlphaCycle.Incr()
         if ss.Stepper.StepPoint(int(AlphaCycle)) {
            return
      }
	}

8. Add code to the user interface to start, pause, and stop the simulation:

   ```go
   func (ss *Sim) ConfigGui() *gi.Window {
      ...
      tbar.AddAction(gi.ActOpts{Label: "Stop", Icon: "stop",
         Tooltip: "Stop the current program at its next natural stopping point (i.e., cleanly stopping when appropriate chunks of computation have completed).",
         UpdateFunc: func(act *gi.Action) {
            act.SetActiveStateUpdt(ss.IsRunning)
      }}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
         fmt.Println("STOP!")
         ss.Stepper.Pause() // NOTE: call Pause here. Stop should only be called when starting over for a new run
         ss.IsRunning = false
         ss.ToolBar.UpdateActions()
         ss.Win.Viewport.SetNeedsFullRender()
      })
      tbar.AddAction(gi.ActOpts{Label: "Cycle", Icon: "run", Tooltip: "Step to the end of a Cycle.",
         UpdateFunc: func(act *gi.Action) {
            act.SetActiveStateUpdt(!ss.IsRunning)
         }}, win.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
         ss.RunSteps(Cycle, tbar)
      })
      ...
   }
   
   func (ss *Sim) RunSteps(grain StepGrain, tbar *gi.ToolBar) {
      if !ss.IsRunning {
         ss.IsRunning = true
         tbar.UpdateActions()
         if ss.Stepper.RunState == stepper.Stopped {
            ss.SimHasRun = true
            ss.OrigSteps = ss.StepsToRun
            ss.Stepper.Start(int(grain), ss.StepsToRun)
            ss.ToolBar.UpdateActions()
            go ss.Train()
      } else if ss.Stepper.RunState == stepper.Paused {
         ss.Stepper.StepGrain = int(grain)
         ss.Stepper.StepsPer = ss.StepsToRun
         ss.Stepper.Enter(stepper.Stepping)
         ss.ToolBar.UpdateActions()
      }
   }

9. Add buttons for selecting a `StepGrain` value for variable-sized steps. See the __PVLV__ model for more detail.

10. That's it!