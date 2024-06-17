# [NE-1690 - Analyse Memory Impact of Pre-Allocated Server Slots for Different Numbers of Routes](https://issues.redhat.com/browse/NE-1690)
HAProxy version 2.8.5-aaba8d0 2023/12/07 - https://haproxy.org/

Column headers:
- ST=<0|1> (server-template disabled=0, enabled=1)
- Tn (Number of Threads)
- RSS (Memory usage in MB)

I have been collecting memory usage data for HAProxy's server-template
feature, where the range is 0..N - this is the ST=1 column.
Additionally, I've collected data where server lines are explicitly
expanded into the haproxy.config for 0..N - this is the ST=0 column.

The memory usage is broadly similar in both cases. However, if you
don't use the server-template, you avoid incurring the memory cost
upfront. When every slot in the server-template is used, it equates to
using the same number of server lines in a backend.

The runtime API provides the ability to add and delete servers
dynamically. Given this API capability, there is no compelling reason
to use the server-template feature.

## Algorithm=leastconn, Weight=1 maxconn=50000
```
#backends  #servers  ST=0 T4 RSS  ST=1 T4 RSS  ST=0 T64 RSS  ST=1 T64 RSS
---------  --------  -----------  -----------  ------------  ------------
      100         0           12           12            17            17
                  1           12           12            18            18
                  2           12           12            19            19
                  5           14           14            21            21
                 10           16           16            25            25
                100           54           54            99            98
                200           97           96           180           179
                300          139          138           262           261
     1000         0           18           18            23            23
                  1           22           22            32            32
                  2           26           26            40            40
                  5           39           39            64            64
                 10           60           60           105           105
                100          444          441           838           835
                200          869          863          1653          1647
                300         1295         1286          2468          2459
    10000         0           80           80            86            86
                  1          123          123           167           168
                  2          166          165           249           249
                  5          293          292           493           492
                 10          506          504           901           898
                100         4338         4308          8234          8204
                200         8595         8534         16382         16322
                300        12852        12761         24531         24440
```
## Algorithm=random, Weight=1 maxconn=50000
```
#backends  #servers  ST=0 T4 RSS  ST=1 T4 RSS  ST=0 T64 RSS  ST=1 T64 RSS
---------  --------  -----------  -----------  ------------  ------------
      100         0           12           12            17            17
                  1           12           12            18            18
                  2           13           13            19            19
                  5           14           14            22            22
                 10           17           17            26            26
                100           62           61           106           106
                200          112          111           195           194
                300          162          161           284           283
     1000         0           18           18            23            23
                  1           23           23            32            32
                  2           28           28            41            41
                  5           43           43            68            68
                 10           68           68           112           112
                100          518          515           913           910
                200         1019         1013          1803          1797
                300         1519         1510          2692          2683
    10000         0           80           80            86            86
                  1          130          131           175           175
                  2          180          181           264           264
                  5          331          330           531           530
                 10          581          578           975           973
                100         5085         5055          8982          8952
                200        10090        10030         17878         17817
                300        15095        15004         26774         26683
```
## Algorithm=roundrobin, Weight=1 maxconn=50000
```
#backends  #servers  ST=0 T4 RSS  ST=1 T4 RSS  ST=0 T64 RSS  ST=1 T64 RSS
---------  --------  -----------  -----------  ------------  ------------
      100         0           12           12            17            17
                  1           12           12            18            18
                  2           13           13            19            19
                  5           14           14            21            21
                 10           16           16            25            25
                100           54           54            99            98
                200           97           96           180           179
                300          139          138           262           261
     1000         0           18           18            23            23
                  1           22           22            32            32
                  2           26           26            40            40
                  5           39           39            64            64
                 10           61           60           105           105
                100          444          441           838           835
                200          869          863          1653          1647
                300         1295         1286          2468          2459
    10000         0           80           80            86            86
                  1          123          123           167           168
                  2          166          166           249           249
                  5          293          292           493           492
                 10          506          504           901           898
                100         4338         4308          8234          8204
                200         8595         8534         16382         16322
                300        12852        12761         24531         24440
```
