package twig

/*
MaxParam URL中最大的参数，注意这个是全局生效的，
无论你有多少路由，请确保最大的参数个数必须小于MaxParam
*/
var MaxParam int = 3

var (
	/*
		NotFoundHandler 全局404处理方法， 如果需要修改
		twig.NotFoundHandler = func (c twig.C) {
				...
		}
	*/
	NotFoundHandler = func(c C) error {
		return ErrNotFound
	}

	/*
		MethodNotAllowedHandler 全局405处理方法
	*/
	MethodNotAllowedHandler = func(c C) error {
		return ErrMethodNotAllowed
	}
)
