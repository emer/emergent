# Copyright (c) 2019, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from emergent import go, gi, giv

# classviews is a dictionary of classviews
classviews = {}

def SetIntValCB(recv, send, sig, data):
    print("set int field:", send)
    vw = gi.SpinBox(send)
    nm = vw.Name()
    print("spin name:", nm)
    nms = nm.split(':')
    cv = classviews[nms[0]]
    flds = cv.Class.__dict__
    setattr(cv.Class, nms[1], vw.Value)

def SetStrValCB(recv, send, sig, data):
    if sig == gi.TextFieldDone:
        print("set text field:", send)
        vw = gi.TextField(send)
        nm = vw.Name()
        nms = nm.split(':')
        cv = classviews[nms[0]]
        flds = cv.Class.__dict__
        setattr(cv.Class, nms[1], vw.Text())

class ClassView(object):
    """
    PyGiClassView provides giv.StructView like editor for python class objects under GoGi
    """
    def __init__(self, name):
        """ note: essential to provide a distinctive name for each view """
        self.Name = name
        classviews[name] = self
        self.Frame = gi.Frame()
        self.Class = None
        
    def SetClass(self, cls):
        self.Class = cls
        self.Config()
        
    def AddFrame(self, par):
        """ Add a new gi.Frame for the view to given parent gi object """
        self.Frame = gi.Frame(par.AddNewChild(gi.KiT_Frame(), "classview"))
    
    def Config(self):
        self.Frame.SetStretchMaxWidth()
        self.Frame.SetStretchMaxHeight()
        self.Frame.Lay = gi.LayoutGrid
        self.Frame.SetPropInt("columns", 2)
        updt = self.Frame.UpdateStart()
        self.Frame.SetFullReRender()
        self.Frame.DeleteChildren(True)
        flds = self.Class.__dict__
        self.Views = {}
        for nm, val in flds.items():
            lbl = gi.Label(self.Frame.AddNewChild(gi.KiT_Label(), "lbl_" + nm))
            lbl.SetText(nm)
            if isinstance(val, int):
                vw = gi.SpinBox(self.Frame.AddNewChild(gi.KiT_SpinBox(), self.Name + ":" + nm))
                vw.SetValue(val)
                vw.SpinBoxSig.Connect(self.Frame, SetIntValCB)
                self.Views[nm] = vw
            else:
                vw = gi.TextField(self.Frame.AddNewChild(gi.KiT_TextField(), self.Name + ":" + nm))
                vw.SetText(str(val))
                vw.TextFieldSig.Connect(self.Frame, SetStrValCB)
                self.Views[nm] = vw
        self.Frame.UpdateEnd(updt)
        
    def Update(self):
        flds = self.Class.__dict__
        for nm, val in flds.items():
            if nm in self.Views:
                vw = self.Views[nm]
                if isinstance(val, int):
                    svw = gi.SpinBox(vw)
                    svw.SetValue(val)
                else:
                    tvw = gi.TextField(vw)
                    tvw.SetText(str(val))


