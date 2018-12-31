package twig

/*
MaxParam URL中最大的参数，注意这个是全局生效的，
无论你有多少路由，请确保最大的参数个数必须小于MaxParam
在你的路由实现中，应该调整这个值到参数的最大值
*/
var MaxParam int = 0
