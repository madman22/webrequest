# webrequest
Helper package for common web request features


A broken example is better than no example in my eyes.  Here's a ~2 year old example that still has to be converted to use the the proper templates and web elements.

```
func (ws *WebServer) Handler(w http.ResponseWriter, r *http.Request) {
	wr := GetWebRequest(r)
	if wr.ID == "" {
		ws.showWebError(errors.New("Missing web request in context"), w, r)
		return
	}
	ws.cachemux.RLock()
	if ws.cache != nil {
		ele, ok := ws.cache[wr.WebRoute]
		ws.cachemux.RUnlock()
		if ok && ele != nil {
			ele.Write(w, r)
			ws.Log(log.Debug, "wrote element from cache")
			return
		}
	} else {
		ws.cachemux.RUnlock()
	}

	var ele dashboard.Element
	var err error
	var template string

	ws.routemux.RLock()
	ele, template, err = ws.routes.Do(wr)
	ws.routemux.RUnlock()
	if err != nil {
		if len(wr.Service) == 0 {
			ws.Log(log.Debug, "Service empty, trying default route")
			ws.defmux.RLock()
			defroute := ws.defroute
			ws.defmux.RUnlock()
			if defroute == nil {
				ws.Log(log.Debug, "no default route, trying ipc")
				f, err := ws.getDefaultFromIpc(wr)
				if err != nil {
					ws.showWebError(err, w, r)
					return
				}
				ws.Log(log.Debug, "got func from ipc")
				ws.defmux.Lock()
				ws.defroute = f
				ws.defmux.Unlock()
				ws.Log(log.Debug, "saved default route")
				defroute = f
			}
			ws.Log(log.Debug, "running default route")
			ele, template, err = defroute(wr)
			if err != nil {
				ws.showWebError(err, w, r)
				return
			}
			ws.Log(log.Debug, "got element from default route")
		} else {
			ws.showWebError(err, w, r)
			return
		}
	}
	if ele == nil {
		ws.showWebError(errors.New("Nil element"), w, r)
		return
	}

	if len(template) > 0 {
		ws.tempmux.RLock()
		temp, ok := ws.templates[template]
		ws.tempmux.RUnlock()
		if ok {
			if !ele.Cacheable() {
				temp.WriteWithChildren(w, r, ele)
				//temp.WriteAndReplace(w, r, ele)
				return
			}
			pg := temp.Copy()
			pg.AppendChildren(ele)
			pg.SetCacheable(true)
			ele = pg
		}
	}

	if ele.Cacheable() {
		ws.Log(log.Debug, "Caching", wr.String())
		ws.cachemux.Lock()
		ws.cache[wr.WebRoute] = ele
		ws.cachemux.Unlock()
	}

	ele.Write(w, r)
}
```
