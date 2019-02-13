// RadixTree Twig中的默认路由
// 参考Echo (https://echo.labstack.com) 中的路由实现
// 做了注释和少量调整

/*
**********************
向Echo开发团队致敬!
向Echo开发者致敬!
向Echo 致敬！
**********************
*/

package twig

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
)

type kind uint8 // 节点类型
type children []*node

type methodHandler struct {
	connect  HandlerFunc
	delete   HandlerFunc
	get      HandlerFunc
	head     HandlerFunc
	options  HandlerFunc
	patch    HandlerFunc
	post     HandlerFunc
	propfind HandlerFunc
	put      HandlerFunc
	trace    HandlerFunc
}

const (
	skind kind = iota // 普通节点
	pkind             // 参数节点 (:xx)
	akind             // 通配节点(*)
)

type node struct {
	kind          kind
	label         byte
	prefix        string
	parent        *node
	children      children
	ppath         string
	pnames        []string
	methodHandler *methodHandler
}

func newNode(t kind, pre string, p *node, c children, mh *methodHandler, ppath string, pnames []string) *node {
	return &node{
		kind:          t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		ppath:         ppath,
		pnames:        pnames,
		methodHandler: mh,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) addHandler(method string, h HandlerFunc) {
	switch method {
	case http.MethodConnect:
		n.methodHandler.connect = h
	case http.MethodDelete:
		n.methodHandler.delete = h
	case http.MethodGet:
		n.methodHandler.get = h
	case http.MethodHead:
		n.methodHandler.head = h
	case http.MethodOptions:
		n.methodHandler.options = h
	case http.MethodPatch:
		n.methodHandler.patch = h
	case http.MethodPost:
		n.methodHandler.post = h
	case PROPFIND:
		n.methodHandler.propfind = h
	case http.MethodPut:
		n.methodHandler.put = h
	case http.MethodTrace:
		n.methodHandler.trace = h
	}
}

func (n *node) findHandler(method string) HandlerFunc {
	switch method {
	case http.MethodConnect:
		return n.methodHandler.connect
	case http.MethodDelete:
		return n.methodHandler.delete
	case http.MethodGet:
		return n.methodHandler.get
	case http.MethodHead:
		return n.methodHandler.head
	case http.MethodOptions:
		return n.methodHandler.options
	case http.MethodPatch:
		return n.methodHandler.patch
	case http.MethodPost:
		return n.methodHandler.post
	case PROPFIND:
		return n.methodHandler.propfind
	case http.MethodPut:
		return n.methodHandler.put
	case http.MethodTrace:
		return n.methodHandler.trace
	default:
		return nil
	}
}

func (n *node) checkMethodNotAllowed() HandlerFunc {
	for _, m := range methods {
		if h := n.findHandler(m); h != nil {
			return MethodNotAllowedHandler
		}
	}
	return NotFoundHandler
}

type radixTreeCtx struct {
	*PureCtx

	handler HandlerFunc
	path    string

	pnames  []string
	pvalues []string

	tree *RadixTree
}

func newRadixTreeCtx(tree *RadixTree) *radixTreeCtx {
	c := &radixTreeCtx{
		PureCtx: NewPureCtx(),
		pvalues: make([]string, tree.maxParam),
		tree:    tree,
		handler: NotFoundHandler,
	}

	return c
}

func (c *radixTreeCtx) Release() {
	c.SetFact(nil)
	c.tree.releaseCtx(c)
}

func (c *radixTreeCtx) Path() string {
	return c.path
}

func (c *radixTreeCtx) Handler() HandlerFunc {
	return c.handler
}

func (c *radixTreeCtx) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}

func (c *radixTreeCtx) Params() UrlParams {
	pms := make(UrlParams)
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			pms[n] = c.pvalues[i]
		}
	}
	return pms
}

func (c *radixTreeCtx) URL(name string, i ...interface{}) (url string) {
	if route, ok := c.tree.routers[name]; ok {
		url = reverse(route.Path(), i...)
	}
	return
}

// RadixTree Twig默认的路由实现，同时也作为一个Wrapper
type RadixTree struct {
	tree    *node
	routers map[string]Router
	m       []MiddlewareFunc
	t       *Twig

	pool     sync.Pool
	maxParam int
}

func NewRadixTree() *RadixTree {
	r := &RadixTree{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routers:  map[string]Router{},
		maxParam: 0,
	}

	r.pool.New = func() interface{} {
		return r.newCtx()
	}

	return r
}

func (r *RadixTree) newCtx() *radixTreeCtx {
	return newRadixTreeCtx(r)
}

func (r *RadixTree) releaseCtx(c *radixTreeCtx) {
	r.pool.Put(c)
}

func (r *RadixTree) Attach(t *Twig) {
	r.t = t
}

func (r *RadixTree) Add(method, path string, h HandlerFunc) {
	if path == "" {
		panic("Twig: path cannot be empty")
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, ppath, pnames)
				return
			}
			r.insert(method, path[:i], nil, pkind, "", nil)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], h, akind, ppath, pnames)
			return
		}
	}

	r.insert(method, path, h, skind, ppath, pnames)
}

func (r *RadixTree) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string) {
	// 调整url最大参数
	l := len(pnames)
	if l > r.maxParam {
		r.maxParam = l
	}

	cn := r.tree // Current node as root
	if cn == nil {
		panic("Twig: invalid method")
	}
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		if l == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.children, cn.methodHandler, cn.ppath, cn.pnames)

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.ppath = ""
			cn.pnames = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, new(methodHandler), ppath, pnames)
				n.addHandler(method, h)
				cn.addChild(n)
			}
		} else if l < sl {
			search = search[l:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			n := newNode(t, search, cn, nil, new(methodHandler), ppath, pnames)
			n.addHandler(method, h)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.ppath = ppath
				if len(cn.pnames) == 0 { // Issue #729
					cn.pnames = pnames
				}
			}
		}
		return
	}
}

func (r *RadixTree) Find(method, path string, ctx *radixTreeCtx) {
	ctx.path = path
	//ctx.SetPath(path)
	cn := r.tree // Current node as root

	var (
		search  = path
		child   *node         // Child node
		n       int           // Param counter
		nk      kind          // Next kind
		nn      *node         // Next node
		ns      string        // Next search
		pvalues = ctx.pvalues // Use the internal slice so the interface can keep the illusion of a dynamic slice
	)

	// Search order static > param > any
	for {
		if search == "" {
			break
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			sl := len(search)
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == akind {
				goto Any
			}
			// Not found
			return
		}

		if search == "" {
			break
		}

		// Static node
		if child = cn.findChild(search[0], skind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' {
				nk = pkind
				nn = cn
				ns = search
			}
			cn = child
			continue
		}

		// Param node
	Param:
		if child = cn.findChildByKind(pkind); child != nil {
			if len(pvalues) == n {
				continue
			}

			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nk = akind
				nn = cn
				ns = search
			}

			cn = child
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Any node
	Any:
		if cn = cn.findChildByKind(akind); cn == nil {
			if nn != nil {
				cn = nn
				nn = cn.parent
				search = ns
				if nk == pkind {
					goto Param
				} else if nk == akind {
					goto Any
				}
			}
			// Not found
			return
		}
		pvalues[len(cn.pnames)-1] = search
		break
	}

	ctx.handler = cn.findHandler(method)
	ctx.path = cn.ppath
	ctx.pnames = cn.pnames

	if ctx.handler == nil {
		ctx.handler = cn.checkMethodNotAllowed()

		if cn = cn.findChildByKind(akind); cn == nil {
			return
		}
		if h := cn.findHandler(method); h != nil {
			ctx.handler = h
		} else {
			ctx.handler = cn.checkMethodNotAllowed()
		}

		ctx.path = cn.ppath
		ctx.pnames = cn.pnames
		pvalues[len(cn.pnames)-1] = ""
	}

	return
}

func (r *RadixTree) Use(m ...MiddlewareFunc) {
	r.m = append(r.m, m...)
}

func (r *RadixTree) Lookup(method, path string, req *http.Request) Ctx {
	c := r.pool.Get().(*radixTreeCtx)
	r.Find(method, path, c)
	c.SetFact(c)
	c.handler = Merge(c.handler, r.m)

	return c
}

func (r *RadixTree) AddHandler(method string, path string, h HandlerFunc, m ...MiddlewareFunc) Router {
	handler := Merge(h, m)
	r.Add(method, path, handler)
	rd := &NamedRoute{
		M: method,
		P: path,
		N: HandlerName(h),
	}
	r.routers[rd.ID()] = rd
	return rd
}

func (r *RadixTree) Match(reqest *http.Request, t *Twig) Lookuper {
	return r
}

func (r *RadixTree) Config() *Config {
	return NewConfig(r)
}

func reverse(path string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' && n < ln {
			for ; i < l && path[i] != '/'; i++ {
			}
			uri.WriteString(fmt.Sprintf("%v", params[n]))
			n++
		}
		if i < l {
			uri.WriteByte(path[i])
		}
	}
	return uri.String()
}
