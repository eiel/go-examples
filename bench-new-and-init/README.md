# 目的

Goで構造体を初期化する場合、どの方法を用いるのが良いのかベンチを取りたい。

# 経緯

Goで構造体を初期化して返す関数を作る人を見かけるが、戻り値をポインタにしているため、戻り値を使うだけでヒープへ確保するように思えた。
必ずヒープへ保存をせざるを得ないのであれば合理的ではあるが、スタックに確保するかどうか選択できる書き方をしたほうが良いのではないかと考えた。

# 結論

まだ検証の余地はあるが、64バイトに収まる構造体であれば、値返しが最も高速
初期化漏れを起こす可能性があるが、ポインタ渡しのメソッドが利便性が高く安定して速い。
ヒープに確保する前提であれば、ポインタ返しでもよい。構造体を直接作られる可能性はあるが初期化漏れも防ぎやすい。

# 想定する書き方

今回想定する書き方は5通り。

- 関数内で構造体を作成、初期化。戻り値を値型で返す(値返し)
- 関数の外で構造体をゼロ値で作成。その後、値レシーバで実装されたメソッドで初期化し、戻り値を値型で返す(値返しメソッド)
- 関数内で構造体を作成、初期化。戻り値はポインタを返す(ポインタ返し)
- 関数の外で構造体をゼロ値で作成。その後、引数にポインタ型で初期化したい構造体を受け取り初期化して、値は返さない(ポインタ渡し)
- 関数の外で構造体をゼロ値で作成。その後、レシーバにポインタ型としたメソッドで初期化。値は返さない(ポインタ渡しメソッド)

## 対象とする構造体


2次元座標を表現する構造体を前提にして進める。

```go
type Point struct {
	x int
	y int
}
```

## 値返し

構造体を返す関数を作る。
今回の検証では、構造体が小さければ最速だった。
CPUアーキテクチャや関数全体のスタックサイズによっても影響を受ける可能性はありそう。

実装例

```go
func NewPoint(x, y int) Point {
	return Point{x: x, y: y}
}
```

利用例

```go
p := NewPoint(1,2)
```

## 値渡し値返しメソッド

構造体を受取、構造体を返す関数を作る。
比較用に用意したパターンで、コピーが2度走るので、絶対に値返しより遅くなってしまう。
どちらかというと、immutableにコピーしたい場合に現れる形かもしれない。


実装例

```go
func (p Point)Init(x, y int) Point {
	p.x = x
	p.y = y
	return p
}
```

利用例

```go
p = Point{}
p.Init(1, 2)
```

## ポインタ返し

値返しとほぼ同じだけど、ポインタ型で返す。
関数のローカルスコープから外れるため、戻り値の値を使うだけでヒープに確保されてしまうようだ。
詳細は確認する必要があるが、今回の方法では構造体にアクセスしているだけアロケーションされており、他のパターンでは起きていない

実装例

```go
func NewPoint(x, y int) *Point {
	return &Point{x: x, y: y}
}
```

利用例

値返しと同じだけど、型がポインタになる点に注意。

```go
p := NewPoint(1,2)
```

## ポインタ渡し

関数の引数に渡すことで、事前に確保したメモリを使えるが、値渡し値返しメソッドのようにコピーが複数回発生してしまう。
なので、事前に確保してポインタで渡す手立てがある。
安定して速い方法になるし、スタックを使うかヒープを使うが使う側で調整も可能。
あえていうなら、利用方法が人によっては不格好に見えるかもしれない。

実装例

```go
func InitPoint(p &Point, x, y int) {
	p.x = x
	p.y = y
}
```

利用例

```go
p = Point{}
InitPoint(&p, 1, 2)
```

利用例 ヒープに確保を強制

```go
p = new(Point)
InitPoint(p, 1, 2)
```

# ポインタ渡しメソッド

ポインタ渡しの不格好さを解決。性能差もない。常用しやすい方法である。

実装例

```go
func (p *Point)Init(x, y int) {
	p.x = x
	p.y = y
}
```

利用例

```
p := &Point{}
p.Init(1, 2)
```

```go
p := new(Point)
p.Init(1, 2)
```


# 実行例

上記の例に近い形で構造体のサイズを大きくさせることで検証した。
今回はint64のフィールドを60個までためした。

実行方法

```
go generate
go test -bench . -benchmem
```


```
goos: darwin
goarch: arm64
pkg: initial
Benchmark_値返し_______________メンバ数1-12                          	1000000000	         0.2924 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数1-12                              	1000000000	         0.2913 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数1-12                             	157535590	         7.714 ns/op	       8 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数1-12                             	1000000000	         0.2899 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数1-12                                 	1000000000	         0.2893 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数2-12                          	1000000000	         0.2898 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数2-12                              	1000000000	         0.2899 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数2-12                             	100000000	        10.34 ns/op	      16 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数2-12                             	1000000000	         0.4341 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数2-12                                 	1000000000	         0.4332 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数3-12                          	1000000000	         0.2893 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数3-12                              	1000000000	         0.2897 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数3-12                             	96089685	        12.25 ns/op	      24 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数3-12                             	1000000000	         0.7231 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数3-12                                 	1000000000	         0.7215 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数4-12                          	1000000000	         0.2886 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数4-12                              	1000000000	         0.2891 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数4-12                             	93020853	        12.39 ns/op	      32 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数4-12                             	1000000000	         0.8660 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数4-12                                 	1000000000	         0.8908 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数5-12                          	1000000000	         1.125 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数5-12                              	345288590	         3.598 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数5-12                             	84640654	        13.37 ns/op	      48 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数5-12                             	1000000000	         1.155 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数5-12                                 	1000000000	         1.165 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数6-12                          	643809781	         1.855 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数6-12                              	315026432	         3.809 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数6-12                             	87704334	        13.74 ns/op	      48 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数6-12                             	924984404	         1.298 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数6-12                                 	921510064	         1.300 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数7-12                          	614694394	         1.962 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数7-12                              	309591099	         3.840 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数7-12                             	83703651	        14.16 ns/op	      64 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数7-12                             	753268086	         1.611 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数7-12                                 	741282022	         1.587 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数8-12                          	473046298	         2.529 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数8-12                              	306087246	         3.967 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数8-12                             	82572354	        14.00 ns/op	      64 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数8-12                             	692605044	         1.733 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数8-12                                 	691393422	         1.732 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数9-12                          	519128908	         2.300 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数9-12                              	298331023	         4.073 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数9-12                             	73274794	        16.45 ns/op	      80 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数9-12                             	592967406	         2.022 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数9-12                                 	593093792	         2.022 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数10-12                         	317947120	         3.779 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数10-12                             	180151810	         6.708 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数10-12                            	72336907	        16.04 ns/op	      80 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数10-12                            	390303716	         2.961 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数10-12                                	384409164	         2.941 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数11-12                         	344527155	         3.474 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数11-12                             	217061271	         5.512 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数11-12                            	64533909	        17.41 ns/op	      96 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数11-12                            	395644232	         2.964 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数11-12                                	398000985	         3.001 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数12-12                         	261431547	         4.507 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数12-12                             	174039136	         6.868 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数12-12                            	67611663	        17.13 ns/op	      96 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数12-12                            	344588412	         3.448 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数12-12                                	331965741	         3.481 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数13-12                         	304804222	         3.931 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数13-12                             	168480256	         7.171 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数13-12                            	63002734	        19.02 ns/op	     112 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数13-12                            	342992869	         3.512 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数13-12                                	352940052	         3.447 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数14-12                         	254067259	         4.769 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数14-12                             	158947738	         7.548 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数14-12                            	62268250	        19.12 ns/op	     112 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数14-12                            	303444058	         3.950 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数14-12                                	304672243	         3.951 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数15-12                         	251785754	         4.798 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数15-12                             	148977362	         7.955 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数15-12                            	59138061	        20.27 ns/op	     128 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数15-12                            	319473861	         3.755 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数15-12                                	317665580	         3.756 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数16-12                         	204883512	         5.934 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数16-12                             	100000000	        10.31 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数16-12                            	56296288	        20.28 ns/op	     128 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数16-12                            	267946930	         4.470 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数16-12                                	268068814	         4.480 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数17-12                         	241699204	         4.957 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数17-12                             	135138666	         8.937 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数17-12                            	55083456	        22.03 ns/op	     144 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数17-12                            	284120000	         4.216 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数17-12                                	282423740	         4.275 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数18-12                         	202128468	         5.925 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数18-12                             	127143442	         9.439 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数18-12                            	52448984	        22.02 ns/op	     144 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数18-12                            	243206310	         4.973 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数18-12                                	238637953	         5.033 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数19-12                         	232735321	         5.144 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数19-12                             	136985017	         8.907 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数19-12                            	52198876	        22.73 ns/op	     160 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数19-12                            	257678989	         4.729 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数19-12                                	249285640	         4.781 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数20-12                         	182826636	         6.554 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数20-12                             	74191429	        16.29 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数20-12                            	52272396	        22.73 ns/op	     160 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数20-12                            	214282923	         5.682 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数20-12                                	211779265	         5.596 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数21-12                         	209489678	         5.742 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数21-12                             	81119678	        14.62 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数21-12                            	49625995	        24.75 ns/op	     176 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数21-12                            	223069645	         5.416 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数21-12                                	219478704	         5.487 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数22-12                         	163578618	         7.332 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数22-12                             	66676852	        17.99 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数22-12                            	44555281	        24.33 ns/op	     176 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数22-12                            	191198089	         6.279 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数22-12                                	190372442	         6.360 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数23-12                         	178373434	         6.701 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数23-12                             	76313022	        15.57 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数23-12                            	46654180	        25.78 ns/op	     192 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数23-12                            	208806198	         5.679 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数23-12                                	211155459	         5.677 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数24-12                         	155578358	         7.724 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数24-12                             	62955224	        19.06 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数24-12                            	43918573	        26.05 ns/op	     192 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数24-12                            	178201941	         6.772 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数24-12                                	178348580	         6.760 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数25-12                         	175540972	         6.829 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数25-12                             	71965096	        16.56 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数25-12                            	43238310	        27.91 ns/op	     208 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数25-12                            	197680507	         6.064 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数25-12                                	197742780	         6.062 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数26-12                         	138969363	         8.637 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数26-12                             	58687239	        20.40 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数26-12                            	41723589	        28.01 ns/op	     208 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数26-12                            	165395029	         7.250 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数26-12                                	164927880	         7.338 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数27-12                         	172062714	         6.984 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数27-12                             	67548393	        17.67 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数27-12                            	40987862	        29.69 ns/op	     224 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数27-12                            	176963364	         6.777 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数27-12                                	177169746	         6.775 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数28-12                         	135980925	         8.911 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数28-12                             	51188352	        21.60 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数28-12                            	39438007	        29.77 ns/op	     224 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数28-12                            	151243838	         7.925 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数28-12                                	151479105	         7.926 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数29-12                         	154711429	         7.648 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数29-12                             	63355045	        18.80 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数29-12                            	38302303	        31.43 ns/op	     240 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数29-12                            	165324010	         7.276 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数29-12                                	165388485	         7.244 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数30-12                         	128643985	         9.292 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数30-12                             	53485968	        22.44 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数30-12                            	38217271	        31.36 ns/op	     240 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数30-12                            	141851613	         8.463 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数30-12                                	141411553	         8.521 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数31-12                         	141116301	         8.493 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数31-12                             	59946296	        19.82 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数31-12                            	36983436	        32.35 ns/op	     256 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数31-12                            	152836339	         7.822 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数31-12                                	151800151	         7.914 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数32-12                         	121398975	         9.882 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数32-12                             	49629673	        24.21 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数32-12                            	35948008	        32.50 ns/op	     256 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数32-12                            	133774472	         9.018 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数32-12                                	132962134	         9.040 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数33-12                         	140293480	         8.578 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数33-12                             	57346641	        20.87 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数33-12                            	32307385	        35.42 ns/op	     288 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数33-12                            	146547164	         8.190 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数33-12                                	144625156	         8.335 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数34-12                         	100000000	        10.92 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数34-12                             	45478804	        26.45 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数34-12                            	33074668	        35.98 ns/op	     288 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数34-12                            	128678852	         9.324 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数34-12                                	128017740	         9.342 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数35-12                         	133275679	         8.988 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数35-12                             	54224626	        22.65 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数35-12                            	32337895	        36.53 ns/op	     288 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数35-12                            	142346720	         8.442 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数35-12                                	140349932	         8.422 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数36-12                         	100000000	        11.12 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数36-12                             	43169996	        27.55 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数36-12                            	31682086	        36.17 ns/op	     288 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数36-12                            	120211160	         9.961 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数36-12                                	120187120	        10.08 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数37-12                         	125512880	         9.495 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数37-12                             	52131508	        23.04 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数37-12                            	31328456	        39.82 ns/op	     320 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数37-12                            	137401512	         8.733 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数37-12                                	136246304	         8.889 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数38-12                         	99762722	        11.57 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数38-12                             	43247658	        27.74 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数38-12                            	29980262	        39.13 ns/op	     320 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数38-12                            	100000000	        10.46 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数38-12                                	100000000	        10.46 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数39-12                         	100000000	        10.49 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数39-12                             	49751758	        23.97 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数39-12                            	29417683	        39.94 ns/op	     320 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数39-12                            	131776129	         9.101 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数39-12                                	131762931	         9.093 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数40-12                         	66936270	        17.81 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数40-12                             	40356877	        29.66 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数40-12                            	29298186	        39.77 ns/op	     320 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数40-12                            	100000000	        11.00 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数40-12                                	100000000	        10.98 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数41-12                         	68721303	        17.41 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数41-12                             	47763332	        25.08 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数41-12                            	27510158	        42.51 ns/op	     352 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数41-12                            	125819392	         9.505 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数41-12                                	125567746	         9.625 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数42-12                         	63610000	        18.78 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数42-12                             	38971267	        31.00 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数42-12                            	26567480	        42.61 ns/op	     352 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数42-12                            	100000000	        11.37 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数42-12                                	100000000	        11.38 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数43-12                         	71841054	        16.75 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数43-12                             	45314296	        26.36 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数43-12                            	26812720	        42.98 ns/op	     352 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数43-12                            	100000000	        10.13 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数43-12                                	120298072	        10.00 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数44-12                         	58294401	        20.69 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数44-12                             	37061200	        32.42 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数44-12                            	26282407	        43.22 ns/op	     352 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数44-12                            	97063159	        11.82 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数44-12                                	100000000	        11.74 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数45-12                         	67178746	        17.77 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数45-12                             	43651045	        27.49 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数45-12                            	25408924	        46.05 ns/op	     384 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数45-12                            	100000000	        10.41 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数45-12                                	100000000	        10.42 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数46-12                         	40238637	        30.23 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数46-12                             	36243054	        33.03 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数46-12                            	24986508	        46.08 ns/op	     384 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数46-12                            	96043864	        12.33 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数46-12                                	97585767	        12.20 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数47-12                         	66864160	        17.98 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数47-12                             	42086126	        28.95 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数47-12                            	25165520	        46.41 ns/op	     384 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数47-12                            	100000000	        10.84 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数47-12                                	100000000	        10.84 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数48-12                         	57816696	        20.74 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数48-12                             	34352250	        34.89 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数48-12                            	25487244	        46.78 ns/op	     384 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数48-12                            	92406318	        12.88 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数48-12                                	93825133	        12.96 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数49-12                         	61876002	        19.28 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数49-12                             	40387891	        29.75 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数49-12                            	23443776	        50.09 ns/op	     416 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数49-12                            	100000000	        11.25 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数49-12                                	100000000	        11.28 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数50-12                         	55103376	        21.81 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数50-12                             	33364032	        35.88 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数50-12                            	22767157	        50.36 ns/op	     416 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数50-12                            	88386394	        13.46 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数50-12                                	88353856	        13.48 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数51-12                         	58317064	        20.52 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数51-12                             	39363297	        30.44 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数51-12                            	22172197	        50.70 ns/op	     416 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数51-12                            	100000000	        11.79 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数51-12                                	98981994	        11.77 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数52-12                         	52161060	        23.02 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数52-12                             	31604486	        37.83 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数52-12                            	23285841	        50.69 ns/op	     416 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数52-12                            	83027503	        14.05 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数52-12                                	85180888	        14.10 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数53-12                         	58150214	        20.66 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数53-12                             	37681243	        31.42 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数53-12                            	21400053	        55.11 ns/op	     448 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数53-12                            	96202663	        12.22 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数53-12                                	96849039	        12.28 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数54-12                         	49899421	        24.11 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数54-12                             	31269408	        38.07 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数54-12                            	21987132	        53.39 ns/op	     448 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数54-12                            	80124416	        14.92 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数54-12                                	80901828	        14.71 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数55-12                         	56234403	        21.21 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数55-12                             	36456064	        32.90 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数55-12                            	21288154	        53.24 ns/op	     448 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数55-12                            	92361566	        12.66 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数55-12                                	93842251	        12.71 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数56-12                         	47538478	        25.45 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数56-12                             	30010659	        39.62 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数56-12                            	22108777	        53.47 ns/op	     448 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数56-12                            	79666726	        15.10 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数56-12                                	79321363	        15.05 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数57-12                         	54587014	        21.96 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数57-12                             	36209972	        33.11 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数57-12                            	20993395	        56.13 ns/op	     480 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数57-12                            	90466183	        13.28 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数57-12                                	90059664	        13.25 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数58-12                         	46536804	        25.77 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数58-12                             	28634862	        41.59 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数58-12                            	20864318	        56.71 ns/op	     480 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数58-12                            	76871133	        15.59 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数58-12                                	77422266	        15.72 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数59-12                         	49383648	        24.90 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数59-12                             	33107786	        35.88 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数59-12                            	20856930	        57.28 ns/op	     480 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数59-12                            	87469135	        13.72 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数59-12                                	86933949	        13.98 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返し_______________メンバ数60-12                         	42722680	        27.11 ns/op	       0 B/op	       0 allocs/op
Benchmark_値返しメソッド_______メンバ数60-12                             	27867066	        43.24 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ返し_________メンバ数60-12                            	20715728	        57.10 ns/op	     480 B/op	       1 allocs/op
Benchmark_ポインタ渡し_________メンバ数60-12                            	74468443	        16.33 ns/op	       0 B/op	       0 allocs/op
Benchmark_ポインタ渡しメソッド_メンバ数60-12                                	74290179	        16.23 ns/op	       0 B/op	       0 allocs/op
PASS
```
