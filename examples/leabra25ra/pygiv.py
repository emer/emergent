# Copyright (c) 2019, The Emergent Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from emergent import go, gi, giv

import pandas as pd

# classviews is a dictionary of classviews -- needed for callbacks
classviews = {}

def SetIntValCB(recv, send, sig, data):
    vw = gi.SpinBox(handle=send)
    nm = vw.Name()
    # print("spin name:", nm)
    nms = nm.split(':')
    cv = classviews[nms[0]]
    flds = cv.Class.__dict__
    setattr(cv.Class, nms[1], vw.Value)

def EditGoObjCB(recv, send, sig, data):
    vw = gi.Action(handle=send)
    nm = vw.Name()
    nms = nm.split(':')
    cv = classviews[nms[0]]
    flds = cv.Class.__dict__
    fld = getattr(cv.Class, nms[1])
    dlg = giv.StructViewDialog(vw.Viewport, fld, giv.DlgOpts(Title=nm), go.nil, go.nil)

def EditObjCB(recv, send, sig, data):
    vw = gi.Action(handle=send)
    nm = vw.Name()
    nms = nm.split(':')
    cv = classviews[nms[0]]
    flds = cv.Class.__dict__
    fld = getattr(cv.Class, nms[1])
    print("editing object: todo: need a ClassViewDialog")
    # dlg = giv.StructViewDialog(vw.Vp, fld.opv.Interface(), DlgOpts{Title: tynm, Prompt: desc, TmpSave: vv.TmpSave}, recv, dlgFunc)    

def SetStrValCB(recv, send, sig, data):
    if sig != gi.TextFieldDone:
        return
    vw = gi.TextField(handle=send)
    nm = vw.Name()
    nms = nm.split(':')
    cv = classviews[nms[0]]
    flds = cv.Class.__dict__
    setattr(cv.Class, nms[1], vw.Text())

def SetBoolValCB(recv, send, sig, data):
    if sig != gi.ButtonToggled:
        return
    vw = gi.CheckBox(handle=send)
    nm = vw.Name()
    # print("cb name:", nm)
    nms = nm.split(':')
    cv = classviews[nms[0]]
    flds = cv.Class.__dict__
    setattr(cv.Class, nms[1], vw.IsChecked() != 0)

class ClassView(object):
    """
    PyGiClassView provides giv.StructView like editor for python class objects under GoGi.
    Due to limitations on calling python callbacks across threads, you must pass a unique
    name to the constructor, along with a dictionary of tags using the same syntax as the struct
    field tags in Go: https://github.com/goki/gi/wiki/Tags for customizing the view properties.
    (space separated, name:"value")
    """
    def __init__(self, name, tags):
        """ note: essential to provide a distinctive name for each view """
        self.Name = name
        classviews[name] = self
        self.Frame = gi.Frame()
        self.Class = None
        self.Tags = tags
        
    def SetClass(self, cls):
        self.Class = cls
        self.Config()
        
    def AddFrame(self, par):
        """ Add a new gi.Frame for the view to given parent gi object """
        self.Frame = gi.Frame(par.AddNewChild(gi.KiT_Frame(), "classview"))
    
    def FieldTags(self, nm):
        """ returns the parsed dictonary of tags for given field """
        tdict = {}
        if nm in self.Tags:
            ts = self.Tags[nm].split(" ")
            for t in ts:
                nv = t.split(":")
                if len(nv) == 2:
                    tdict[nv[0]] = nv[1].strip('"')
                else:
                    print("ClassView: error in tag formatting for field:", nm, 'should be name:"value", is:', t)
        return tdict

    def HasTagValue(self, tags, tag, value):
        """ returns true if given tag has given value """
        if not tag in tags:
            return False
        tv = tags[tag]
        return tv == value
        
    def Config(self):
        self.Frame.SetStretchMaxWidth()
        self.Frame.SetStretchMaxHeight()
        self.Frame.Lay = gi.LayoutGrid
        self.Frame.Stripes = gi.RowStripes
        self.Frame.SetPropInt("columns", 2)
        updt = self.Frame.UpdateStart()
        self.Frame.SetFullReRender()
        self.Frame.DeleteChildren(True)
        flds = self.Class.__dict__
        self.Views = {}
        for nm, val in flds.items():
            tags = self.FieldTags(nm)
            if self.HasTagValue(tags, "view", "-"):
                continue
            lbl = gi.Label(self.Frame.AddNewChild(gi.KiT_Label(), "lbl_" + nm))
            lbl.SetText(nm)
            if isinstance(val, bool):
                vw = gi.CheckBox(self.Frame.AddNewChild(gi.KiT_CheckBox(), self.Name + ":" + nm))
                vw.SetChecked(val)
                vw.ButtonSig.Connect(self.Frame, SetBoolValCB)
                if self.HasTagValue(tags, "inactive", "+"):
                    vw.SetInactive()
                self.Views[nm] = vw
            elif isinstance(val, go.GoClass):
                vw = gi.Action(self.Frame.AddNewChild(gi.KiT_Action(), self.Name + ":" + nm))
                if hasattr(val, "Label"):
                    vw.SetText(val.Label())
                else:
                    vw.SetText(nm)
                vw.SetPropStr("padding", "2px")
                vw.SetPropStr("margin", "2px")
                vw.SetPropStr("border-radius", "4px")
                vw.ActionSig.Connect(self.Frame, EditGoObjCB)
                if self.HasTagValue(tags, "inactive", "+"):
                    vw.SetInactive()
                self.Views[nm] = vw
            elif isinstance(val, pd.DataFrame):
                vw = gi.Action(self.Frame.AddNewChild(gi.KiT_Action(), self.Name + ":" + nm))
                vw.SetText(nm)
                vw.SetPropStr("padding", "2px")
                vw.SetPropStr("margin", "2px")
                vw.SetPropStr("border-radius", "4px")
                vw.ActionSig.Connect(self.Frame, EditObjCB)
                if self.HasTagValue(tags, "inactive", "+"):
                    vw.SetInactive()
                self.Views[nm] = vw
            elif isinstance(val, (int, float)):
                vw = gi.SpinBox(self.Frame.AddNewChild(gi.KiT_SpinBox(), self.Name + ":" + nm))
                vw.SetValue(val)
                vw.SpinBoxSig.Connect(self.Frame, SetIntValCB)
                if self.HasTagValue(tags, "inactive", "+"):
                    vw.SetInactive()
                self.Views[nm] = vw
            else:
                vw = gi.TextField(self.Frame.AddNewChild(gi.KiT_TextField(), self.Name + ":" + nm))
                vw.SetText(str(val))
                vw.TextFieldSig.Connect(self.Frame, SetStrValCB)
                if self.HasTagValue(tags, "inactive", "+"):
                    vw.SetInactive()
                self.Views[nm] = vw
        self.Frame.UpdateEnd(updt)
        
    def Update(self):
        flds = self.Class.__dict__
        for nm, val in flds.items():
            if nm in self.Views:
                vw = self.Views[nm]
                # print("updating:", nm, "view:", vw)
                if isinstance(val, bool):
                    svw = gi.CheckBox(vw)
                    svw.SetChecked(val)
                elif isinstance(val, go.GoClass):
                    pass
                elif isinstance(val, pd.DataFrame):
                    pass
                elif isinstance(val, (int, float)):
                    svw = gi.SpinBox(vw)
                    svw.SetValue(val)
                else:
                    tvw = gi.TextField(vw)
                    tvw.SetText(str(val))


