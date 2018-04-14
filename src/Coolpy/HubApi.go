package Coolpy

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"fmt"
	"strconv"
)

func HubPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	//post接口允许模拟put提交
	//hub节点put api/hubs?method=put&hid=1
	qs := r.URL.Query()
	if qs.Get("method") == "put" {
		if qs.Get("hid") != "" {
			nps := append(ps, httprouter.Param{"hid", qs.Get("hid")})
			HubPut(w, r, nps)
			return
		}
	}
	decoder := json.NewDecoder(r.Body)
	var h Hub
	err := decoder.Decode(&h)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, err)
		return
	}
	_, err = r.Cookie("islogin")
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "dosn't login")
		return
	}
	errs := CpValidate.Struct(h)
	if errs != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "invalid")
		return
	}
	uc, _ := r.Cookie("ukey")
	h.Ukey = uc.Value
	err = hubCreate(&h)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, err)
		return
	}
	pStr, _ := json.Marshal(&h)
	fmt.Fprintf(w, `{"ok":%d,"data":%v}`, 1, string(pStr))
}

func HubsGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	_, err := r.Cookie("islogin")
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "dosn't login")
		return
	}
	ukey, _ := r.Cookie("ukey")
	ndata, err := hubStartWith(ukey.Value)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, err)
		return
	}
	pStr, _ := json.Marshal(&ndata)
	fmt.Fprintf(w, `{"ok":%d,"data":%v}`, 1, string(pStr))
}

type RHub struct {
	Id     uint64
	Title  string
	About  string
	Tags   []string
	Public bool
	RNodes []*RNode
}

type RNode struct {
	Id        uint64
	Title     string
	About     string
	Tags      []string
	Type      int
	CtrlerVal string
}

func HubsAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	_, err := r.Cookie("islogin")
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "dosn't login")
		return
	}
	ukey, _ := r.Cookie("ukey")
	ndata, err := hubStartWith(ukey.Value)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, err)
		return
	}
	var Rhub []*RHub
	for _, h := range ndata {
		nhub := &RHub{Id:h.Id, Title:h.Title, About:h.About, Tags:h.Tags, Public:h.Public}
		strid := strconv.FormatUint(h.Id, 10)
		key := ukey.Value + ":" + strid + ":"
		nodes, _ := NodeStartWith(key)
		for _, n := range nodes {
			nnode := &RNode{Id:n.Id, Title:n.Title, About:n.About, Tags:n.Tags, Type:n.Type}
			if n.Type == NodeTypeEnum.Switcher {
				sws, _ := GetSwitcher(key + strconv.FormatUint(n.Id, 10))
				nnode.CtrlerVal = strconv.Itoa(sws.Svalue)
			} else if n.Type == NodeTypeEnum.GenControl {
				gen, _ := GetGenControl(key + strconv.FormatUint(n.Id, 10))
				nnode.CtrlerVal = gen.Gvalue
			} else if n.Type == NodeTypeEnum.RangeControl {
				ran, _ := GetRangeControl(key + strconv.FormatUint(n.Id, 10))
				//val,min,max,step
				nnode.CtrlerVal = strconv.FormatInt(ran.Rvalue, 10) + "," + strconv.FormatInt(ran.Min, 10) + "," + strconv.FormatInt(ran.Max, 10) + "," + strconv.FormatInt(ran.Step, 10)
			}
			nhub.RNodes = append(nhub.RNodes, nnode)
		}
		Rhub = append(Rhub, nhub)
	}
	pStr, _ := json.Marshal(&Rhub)
	fmt.Fprintf(w, `{"ok":%d,"data":%v}`, 1, string(pStr))
}

func HubGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	//get接口允许模拟delete提交
	//hub节点put api/hub/:hid?method=delete
	qs := r.URL.Query()
	if qs.Get("method") == "delete" {
		HubDel(w, r, ps)
		return
	}
	hid := ps.ByName("hid")
	if hid == "" {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "params ukey")
		return
	}
	_, err := r.Cookie("islogin")
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "dosn't login")
		return
	}
	ukey, _ := r.Cookie("ukey")
	ndata, err := HubGetOne(ukey.Value + ":" + hid)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "hub not ext")
		return
	}
	pStr, _ := json.Marshal(&ndata)
	fmt.Fprintf(w, `{"ok":%d,"data":%v}`, 1, string(pStr))
}

func HubPut(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	hid := ps.ByName("hid")
	if hid == "" {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "params ukey")
		return
	}
	decoder := json.NewDecoder(r.Body)
	var h Hub
	err := decoder.Decode(&h)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, err)
		return
	}
	_, err = r.Cookie("islogin")
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "dosn't login")
		return
	}
	ukey, _ := r.Cookie("ukey")
	oh, err := HubGetOne(ukey.Value + ":" + hid)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "hub nrole")
		return
	}
	oh.About = h.About
	//oh.Latitude = h.Latitude
	//oh.Local = h.Local
	//oh.Longitude = h.Longitude
	oh.Public = h.Public
	oh.Tags = h.Tags
	oh.Title = h.Title
	hubReplace(ukey.Value + ":" + hid, oh)
	pStr, _ := json.Marshal(&oh)
	fmt.Fprintf(w, `{"ok":%d,"data":%v}`, 1, string(pStr))
}

func HubDel(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	defer r.Body.Close()
	hid := ps.ByName("hid")
	if hid == "" {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "params ukey")
		return
	}
	_, err := r.Cookie("islogin")
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "dosn't login")
		return
	}
	ukey, _ := r.Cookie("ukey")
	key := ukey.Value + ":" + hid
	_, err = HubGetOne(key)
	if err != nil {
		fmt.Fprintf(w, `{"ok":%d,"err":"%v"}`, 0, "hub not ext")
		return
	}
	//delete all sub node
	hubdel(key)
	fmt.Fprintf(w, `{"ok":%d}`, 1)
}