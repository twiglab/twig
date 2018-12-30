package twig

import (
	"net/http"
)

/*
Muxer 接口
*/
type Muxer interface {
	Lookup(string, string, *http.Request, Ctx)
	Add(string, string, HandlerFunc, ...MiddlewareFunc) *Route
	Attacher
}

/*
RadixTreeMux 基于Radix树实现路由
*/

type kind uint8 // Radix Tree的节点类型
type children []*node

// methodHandler 位于节点中，用于存放HandlerFunc
// 根据method 用switch 查找
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
	skind kind = iota // 可以忽略的节点(普通节点)
	pkind             // 参数节点 (:xxx)
	akind             // * 节点
)

type node struct {
	kind          kind           // 节点的类型
	label         byte           // 对应的Radix的label ， 就是当前前缀的第一个字符(参考RadixTree实现)
	prefix        string         // 节点的前缀
	parent        *node          //上级节点
	children      children       // 子节点
	ppath         string         // 当前节点对应的路径（就是URL，带参数形式）
	pnames        []string       //如果是参数节点，保存参数列表
	methodHandler *methodHandler // 当前节点对应的handler
}

func newNode(t kind, pre string, p *node, c children, mh *methodHandler, ppath string, pnames []string) *node {
	return &node{
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

// findChild 从子节点中查找符合一起的接单
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

// addHandler 放置HandlerFunc
// 小量数据switch 效率优于map
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

// RadixTree 路由实现
type RadixTreeMux struct {
	tree            *node
	mid             []MiddlewareFunc //路由级中间件
	NotFoundHandler HandlerFunc

	t *Twig

	routes map[string]*Route
}

func NewRadixTreeMux() *RadixTreeMux {
	return &RadixTreeMux{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		NotFoundHandler: NotFoundHandler, //路由级别404
		routes:          make(map[string]*Route),
	}
}

func (r *RadixTreeMux) Attach(t *Twig) {
	r.t = t
}

func (r *RadixTreeMux) Use(m ...MiddlewareFunc) {
	r.mid = append(r.mid, m...)
}

func (r *RadixTreeMux) Add(method, path string, handler HandlerFunc, m ...MiddlewareFunc) *Route {
	if path == "" {
		panic("twig: path cannot be empty")
	}
	if path[0] != '/' {
		path = "/" + path
	}

	// HandlerFunc 级Middleware
	h := Enhance(handler, m)
	pnames := []string{} // 参数列表
	ppath := path        // path

	// 一次扫描，处理参数列表，生成RadixTree
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' { //参数处理
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil) // 用参数前面的路径构建节点(普通节点)
			for ; i < l && path[i] != '/'; i++ {
			}

			// 上面的循环是跳过参数，并记录位置
			pnames = append(pnames, path[j:i]) // 肯定是参数了
			path = path[:j] + path[i:]         // 去掉参数，后续接着扫描
			i, l = j, len(path)                //记录去掉当前参数后的path的位置

			if i == l { // 到结尾了，肯定是带参数，但不以/结束的，例如 /:id 这种路由
				r.insert(method, path[:i], h, pkind, ppath, pnames) //构建一个参数节点，结束
				return r.addRoute(method, path, handler)
			}
			r.insert(method, path[:i], nil, pkind, "", nil) // 还没结算
		} else if path[i] == '*' { // 通配符(*)处理, 后面不用在扫描了
			r.insert(method, path[:i], nil, skind, "", nil)
			pnames = append(pnames, "*") // 通配符的参数名称就是 *
			r.insert(method, path[:i+1], h, akind, ppath, pnames)
			return r.addRoute(method, path, handler)
		}
	}

	r.insert(method, path, h, skind, ppath, pnames) // 整个路由都没有参数，就是一个普通的节点

	return r.addRoute(method, path, handler)
}

func (r *RadixTreeMux) addRoute(method, path string, handler HandlerFunc) *Route {
	route := &Route{
		Method: method,
		Path:   path,
		Name:   HandlerName(handler),
	}
	r.routes[method+path] = route
	return route
}

func (r *RadixTreeMux) insert(method, path string, h HandlerFunc, t kind, ppath string, pnames []string) {
	// 调整url最大参数
	// 优化: MaxParam为*整个*服务器中最大的路由参数个数，后续分配Ctx时，按照最大参数分配参数值
	l := len(pnames)
	if l > MaxParam {
		MaxParam = l
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
				if len(cn.pnames) == 0 {
					cn.pnames = pnames
				}
			}
		}
		return
	}
}

func (r *RadixTreeMux) Lookup(method, path string, req *http.Request, ctx Ctx) {
	ctx.SetPath(path)
	cn := r.tree // Current node as root

	var (
		search  = path
		child   *node               // Child node
		n       int                 // Param counter
		nk      kind                // Next kind
		nn      *node               // Next node
		ns      string              // Next search
		pvalues = ctx.ParamValues() // Cached ...
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
			if cn.prefix[len(cn.prefix)-1] == '/' {
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

	ctx.SetHandler(cn.findHandler(method))
	ctx.SetPath(cn.ppath)
	ctx.SetParamNames(cn.pnames)

	if ctx.Handler() == nil {
		ctx.SetHandler(cn.checkMethodNotAllowed())

		if cn = cn.findChildByKind(akind); cn == nil {
			return
		}
		if h := cn.findHandler(method); h != nil {
			ctx.SetHandler(h)
		} else {
			ctx.SetHandler(cn.checkMethodNotAllowed())
		}
		ctx.SetPath(cn.ppath)
		ctx.SetParamNames(cn.pnames)
		pvalues[len(cn.pnames)-1] = ""
	}

	return
}
