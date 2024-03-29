package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gitee.com/dark.H/go-remote-repl/datas"
	"gitee.com/dark.H/go-remote-repl/remote"
)

var (
	// PASWD = "dGxzOi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlFbVRDQ0E0R2dBd0lCQWdLQ0FRQlZGbm1Ja3lhVGcrUXR3S2d6aHhjVE9MdDlvZnl3alNMS01XL2YwSlFQCm1md0x1b3gzc08yc0hKK2FHaHhCYTVINkhOaXl5TitTMmlkck55OVNnRi80WiswMUl2c29NbVBDVzdwblhrZTgKQXg4LzVUTWJUV3VXM0FxamRTZHZXdjlIMkI1NExYSXk5WkJVdUl3U3pJMW90cmZrNFh1N2xNa01sS0Q1dk9IQgpxWWtOVDY4cEl6VTVNTFZoWVN4bUNSRnI1dkhvbXJYWUVCejBPTU1ZcHlqRjBrajNzZks1Zlh3NDdsclo1VnpPClFOZURhajIrZWlaUWVNWWYvaWVHTHpza0VPRUYvdUc4dGxaN052RUUxdTdhd0doRmJTQy9xYjhYeWk5b3d5ME4KU2g2Z1ZibUp2SFBrZzVFT1lZRm5lS1ZDZnE0WXpRWnZvVUVzN0drRWZQQ3RNQTBHQ1NxR1NJYjNEUUVCQ3dVQQpNR1V4Q3pBSkJnTlZCQVlUQWtwUU1RNHdEQVlEVlFRSUV3VlViMnQ1YnpFT01Bd0dBMVVFQnhNRlZHOXJlVzh4CkVEQU9CZ05WQkFrVEIxUmhjMk5wYTI4eEVUQVBCZ05WQkJFVENERXdMVEl3TUMwME1SRXdEd1lEVlFRS0V3aEwKWTNCbFpTQkRiekFlRncweU1EQTNNRGd3T0RNM016SmFGdzB5TVRBM01EZ3dPRE0zTXpKYU1HVXhDekFKQmdOVgpCQVlUQWtwUU1RNHdEQVlEVlFRSUV3VlViMnQ1YnpFT01Bd0dBMVVFQnhNRlZHOXJlVzh4RURBT0JnTlZCQWtUCkIxUmhjMk5wYTI4eEVUQVBCZ05WQkJFVENERXdMVEl3TUMwME1SRXdEd1lEVlFRS0V3aExZM0JsWlNCRGJ6Q0MKQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFMdFNHdXVLWGhzVWY0cDNzT2I2QnNBNAo3WCtDRVpJSys2RzZiWFp4V2poSDdCNDI0amdPNS9XZ0ZDbzNySDEwK0FvRU5BRGtBcUUrYjAyTWgzb0d4Wm9QCk5yT0hXeE5Rb2puVlBVd1drMW05NGEra2JIcHNQdHdDd0FnNzBmbTJlNUJKK3ZMNEJEdUsyeEZta1Zub2dpN0kKMDJFWmtDcFdQaU5kMTlaMVdvWmY4MlVCaFBoQ3RyRURGV1NWeW1Vc0lac1Rsa2JWbmJXQzhnMU1Wa2xJdlMxNgp4aGVHa1kyR2hNOEViWUJrZmZ1bzlyT2Fmazh3dTlKT0V5eTFzelIxL3FMZjA0Q0xGa0dMdGRKbnFrSmJaQ2hYCllTRXZqSEpzVkRhMDg1ODZzUGZPckQrRUlsODlyZ1JPalJuWDFGWlBiVW41OW8xN05XTWFTTmtINVdaOU5Ic0MKQXdFQUFhTlRNRkV3RGdZRFZSMFBBUUgvQkFRREFnV2dNQjBHQTFVZEpRUVdNQlFHQ0NzR0FRVUZCd01CQmdncgpCZ0VGQlFjREFqQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01BOEdBMVVkRVFRSU1BYUhCSDhBQUFFd0RRWUpLb1pJCmh2Y05BUUVMQlFBRGdnRUJBSnZsODd0RXVQZ3labGZma3BNMzFtK2ZjN2pralUzdjVla0pwWUsreVVDS3B2d0sKSGo3V3ZTdW1oa0cyb0VHTGdKeUh1WEdtT2E3NFp4RWhSY3RRd3VFODM4M0NhYXkrZmczeXQ0NGZPa0ExVzhRSwpKZXQ3K21HeERESVJQM0s3VExvMWdmS09kRFJ5RTc1N0NMeG1mcHNkbngwbWRUQUIyWDgwM05rTmk3TXNka3NPCm5QTmhjQjZyVDc4Q0RTSlpwUkZvK1RTMytEUVQ0c0daT05vSkN1OVhuQ0VPaFk2c0xBb1FxVVl5d0xibVJFaG4KUThXbHFLN1BnTDR3SmViNHRieEwyeWhoelI3TDQ3VGtrb2w1Z0ZNRGRQcGxQZVNSbWdCaG0vSnJaRkN3WFZDUAp2WWVhQlNmdkxid0hUaWVaL1VBVGlyaUVUaGg2WndRVHM5KytYMlE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KPFNFUD4tLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS0KTUlJRXZnSUJBREFOQmdrcWhraUc5dzBCQVFFRkFBU0NCS2d3Z2dTa0FnRUFBb0lCQVFDN1VocnJpbDRiRkgrSwpkN0RtK2diQU9PMS9naEdTQ3Z1aHVtMTJjVm80Uit3ZU51STREdWYxb0JRcU42eDlkUGdLQkRRQTVBS2hQbTlOCmpJZDZCc1dhRHphemgxc1RVS0k1MVQxTUZwTlp2ZUd2cEd4NmJEN2NBc0FJTzlINXRudVFTZnJ5K0FRN2l0c1IKWnBGWjZJSXV5Tk5oR1pBcVZqNGpYZGZXZFZxR1gvTmxBWVQ0UXJheEF4VmtsY3BsTENHYkU1WkcxWjIxZ3ZJTgpURlpKU0wwdGVzWVhocEdOaG9UUEJHMkFaSDM3cVBhem1uNVBNTHZTVGhNc3RiTTBkZjZpMzlPQWl4WkJpN1hTClo2cENXMlFvVjJFaEw0eHliRlEydFBPZk9yRDN6cXcvaENKZlBhNEVUbzBaMTlSV1QyMUorZmFOZXpWakdraloKQitWbWZUUjdBZ01CQUFFQ2dnRUJBSUV2SFl1bFVFTWl2dGF5dHRpRVE5V2JMUWRMbjJ4MWtKWTNxTGdQY21YdwpEQUhHWGZyRkpPbDdiWXl6K294Rzk2eUl1NGlGdHQ2VzNWSEJBY21CRkJPc1BGQlJuOWpST2gvaVZMUzl1UDhoCitad1ZMQTY5eWhnVytYSjBOdG9kSkZnYy9TbEZMRDRZQS82Ykd0RERqY3hRQ1o4a2J3MFRqeVQ2MmhoM3RHZUkKWkkvN2RCeUFTSk1nODZRdlJzaDdiOTJYYVZ3eXM0SHpDcVNDeGVrZDIxVmVSWlQzRnNVSXUzbkptamNZazJVdgo3OWdUQWt1aUNydnpBcG81TzBybnphSEJhUEl6YThUUWdCamJxMnpNT1hCOGsyUktGZUNCY2RId1U5UkIzUlVHCmE1cjhLSFhrS09pK1JKT2s1b0Z1d05hRUo1UkcyVlY2c2lSYTE0ZlM4REVDZ1lFQTNwektlSFpCUzJLVnpURSsKOVR4citTVVBxZlkvZnNNRk9YdEZYSlpkRmhQQnRkSkk5MEllL2N3UDVYeE16VFdzeHBPOENIdVNocU1BbUZyagpGa1pFdWVFQlcwWnVhellGT3dSdkd1YUlsU05RWjdoVHZSRWJ6UXFVSjFwMzBXWjE2LzQyNDMyY0F0NXhTTXUxCmRoZVB1bzF6VFNnbWZmV2pjWFM0SXdkeWg3a0NnWUVBMTJwSmpCanlHbUFlVjhhcHB6Z1dkYm1IblJDazVQNkUKdmJpN3Vid2tta2FVUDByY2ZOaWl2aHB0UUFidWNvMytZVHRyZ3RWYTN2WFkrTTkwd0JxS3NMTk9jR21xWnFIOQppNHVzanI2TGdJYzlMaWo1N2tleklHOVU3UWJDRllaSUtFU24va0pyODVCYW14QUh5cGQyeTB0ZWF2UFI4TU5xCnFteWdaTmovajlNQ2dZRUF3S1Y1Um5RNEgwL3FpWTlqaDRESmcrdkJ1WGtrQzhRam9sSCtmWmlXYUFPaThJRlIKcWNDcjUwRVFSSzJrSFRhbEZaaEY4aVZXY1hOZ2tYaWQ2dW5Xa2ZHWlplNzJvWXMzVFpLUnYvcmZHZ2xjak5YawppY3JqZnpiM2JYTUtHOG9LcExienN6MUJwTzN4cFdpODJieWNJbnBFM1BHeEZmSmNobFBYQW1Gd2lPa0NnWUFRCldqb3hEMmU1aHRobTAyYm5rY05xdG1YTUQ0cGs4OGxCMmJjdWQxRFVBVTJackdZbWVBK0tuTmowUmxKdGtjZmcKdTdFQ29vMmVaVTFPUGZsZjUrUGxYQWMxVTJBaklHTHQ1L0YyZUpLQWRFTzVzRlNndVFLcEVLcUN2OE9WL0xhWApqL2FzdjRBUVlrSHVpWDM5N0JlUmdTd3V0RW1ZbkUwSm1PdG9IR3p5RHdLQmdDNTBtUDNVbzVYUmpYWlQzNnBsClU5SE51UzQ2OW8wcWNFd2d1b1JvUWxUWXp1MXJkT0M1WCtGbEpaWk5kcGczWVVUR0t6cXhGTVpLOVRjNzdnelMKRDV6VWVTQlBhamZiTVJmVHRsZkp2MmhBTEFBeEFWc3FMS3MzVkZPNDM2dGZyVFpQRlZNYWZpRU9OWkFLZ3VxaQpYcFY1OU5DTFdsN1RBUVFSWFRSZDdPNXIKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQpAMTI3LjAuMC4xOjEyMzQ1"
	PASWD = "dGxzOi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlFdURDQ0E2Q2dBd0lCQWdLQ0FRQXVsSEcvUHNKUmVIVWJuelh3SVU2MDlOb1lBZVU4OXFLOVp1WVpZZllrCjVJRHNGZnZlaWV6aGcvM0tndTl3cVdsZFdreHgxMjV6LzdNOXQrMDI2T0c2WkZFTG5sb0U2UmZ2K1RFMmV0bC8Kd3VsZTh2Q0xhdHZRUVp4cnhxQ1FxZHpEZTZiUmsralgyQ2dVRzR2Qy9wUEJFaXZVZ2RTbUJEUTRoVFRiSzlVOQpWbm04QkhqV0JkMVJOdU1BYk9iekJCd0RTZmZJL1RzVWVTUHhHcXA2MFJuZXJxbXFpSWloUHpldVdrcXBZbjBvCk4rTFlHb2hmZGdjYnREeVVVZ0xIelc1RFJFdndDblpORnVNWkJxdEp0N2tkbE5xeWkxaHMvaUk0LzZEV3NBQUYKanQ1UnkySHNXN2NhQm1Ma2lIVjh2aWdxOEh3SXVwbDVVdmpXSXNuT3pLcytNQTBHQ1NxR1NJYjNEUUVCQ3dVQQpNR1V4Q3pBSkJnTlZCQVlUQWtwUU1RNHdEQVlEVlFRSUV3VlViMnQ1YnpFT01Bd0dBMVVFQnhNRlZHOXJlVzh4CkVEQU9CZ05WQkFrVEIxUmhjMk5wYTI4eEVUQVBCZ05WQkJFVENERXdMVEl3TUMwME1SRXdEd1lEVlFRS0V3aEwKWTNCbFpTQkRiekFlRncweU1URXlNakF3TlRNMU1qWmFGdzB5TWpFeU1qQXdOVE0xTWpaYU1HVXhDekFKQmdOVgpCQVlUQWtwUU1RNHdEQVlEVlFRSUV3VlViMnQ1YnpFT01Bd0dBMVVFQnhNRlZHOXJlVzh4RURBT0JnTlZCQWtUCkIxUmhjMk5wYTI4eEVUQVBCZ05WQkJFVENERXdMVEl3TUMwME1SRXdEd1lEVlFRS0V3aExZM0JsWlNCRGJ6Q0MKQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFQVlErSnY1NStDN2RHbzNxYy9HYjk4SwpBNmJEdG1mK0VzT28wb1paUEkwdmZKN3ExY1YyM2dzNkc4UmhPaFVoWG12UEJCdFZHVGpXbTZ3NjJtNDFBcjNhCitYNzRwcjVJcHpubkp4U0RsdG9IQXBtcThyR3JTZE1HTkQ5czhuNE9HdXgvdUtuTjVsTUhaTnVjTHFuWnl1VVAKdHE4czBEL1dWOGg4QS9iU3RLWEhPTU1wSkRobUExUlNiWHVaZjY3RWhVOG9ER3ZISDdaVHoySHNnR1ovL2t4KwpZdTdXY0YzVlhHdUlub3Jmb0RFUm9lTkQ5bCtPSGo5UW9sV0cyRmVValIzRGYwZFpEcmVvNVprdHNPbVFadDFtCjRDa0NVRWZaU1NRK1J5Tmc4dnBxQmNSaUlkcWRwNXBSZVhyUWdGVndmWDJGMnVwMDJ2V1VGYVhiVEJiL1RUTUMKQXdFQUFhTnlNSEF3RGdZRFZSMFBBUUgvQkFRREFnV2dNQjBHQTFVZEpRUVdNQlFHQ0NzR0FRVUZCd01CQmdncgpCZ0VGQlFjREFqQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCU1RUcHhJc0xnQm44UXRVUzhYCko4VjNnT28xSXpBUEJnTlZIUkVFQ0RBR2h3UUFBQUFBTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDUnJxenEKcXJWWWIwbm9SU3BrbkN0U1hpeHViczZLRDJPTWN3VUNoU0hZYjdWNE1YTVFWblNRLzdkeG9LR0o1b1E0UWo4dwplaHFyY3NMTWRBS0lNcXN5dm1ZTXpURFdkdy8ya2NsWHROYm9pTkFXTEV1NkJucXZ4TERlQ1JVSW1sck9haTd2CnV1a250U2tZT1Irei9EWVNqK2xmRERsbHl5bmJNZmt6OGlpK3d3NDBQamtjUVdKeFFBR3MwcGJjOVRTNGdjK1cKRmhpSXcxS2R5VFd3NjhNaW5lalV1MElYVHlSMlhoSFBsOUl6VTRVbyttdlRKWitOQVJ3K3g4SkhNcUxWUGx2aAovRUw3SS9jYVQ4VFVSRmxWNjAzYUtrQkV1OStEVWFHMjNxUDZXdTVGa1dEMzNHTVN3eGpLeENLclR5a003NjFUCm9kMFloRSthMHFkMkpDZC8KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo8U0VQPi0tLS0tQkVHSU4gUFJJVkFURSBLRVktLS0tLQpNSUlFdkFJQkFEQU5CZ2txaGtpRzl3MEJBUUVGQUFTQ0JLWXdnZ1NpQWdFQUFvSUJBUUQxVVBpYitlZmd1M1JxCk42blB4bS9mQ2dPbXc3Wm4vaExEcU5LR1dUeU5MM3llNnRYRmR0NExPaHZFWVRvVklWNXJ6d1FiVlJrNDFwdXMKT3RwdU5RSzkydmwrK0thK1NLYzU1eWNVZzViYUJ3S1pxdkt4cTBuVEJqUS9iUEorRGhyc2Y3aXB6ZVpUQjJUYgpuQzZwMmNybEQ3YXZMTkEvMWxmSWZBUDIwclNseHpqREtTUTRaZ05VVW0xN21YK3V4SVZQS0F4cnh4KzJVODloCjdJQm1mLzVNZm1MdTFuQmQxVnhyaUo2SzM2QXhFYUhqUS9aZmpoNC9VS0pWaHRoWGxJMGR3MzlIV1E2M3FPV1oKTGJEcGtHYmRadUFwQWxCSDJVa2tQa2NqWVBMNmFnWEVZaUhhbmFlYVVYbDYwSUJWY0gxOWhkcnFkTnIxbEJXbAoyMHdXLzAwekFnTUJBQUVDZ2dFQUpyUlR4U2cwc1h4aDdiQkNsb21VZFBnN2cwQTFzMHFlNDBmcmowN2RBdG1QCksvT1pMc0tHWGZjWkVtN1VueWpZb09IVW9BdFpJMktZeFIrei90ZFdTMllWSHpkVDk0bTIrS2ZEdnJUVVJDbisKK3dmQnFzNS9vMC9EUm82RjQrV04wZXFFMUNXc1JVVlJZR1RQNS9SNDBwTWNHNTlTYkFBQmNady9jOTNpcXBUcApCbllRdVAxemVaSFF0ZXphRXcyL1o0TE5sRjFNTmtlMkthanlEVm4vMVE4YnNzblhGVk9tUWpaUnVuaUQxcEloCkcyL04rNTBJWTB4d0F0am5vZEtvVGxWdDFmYzVWWXRDZnBRODVMR0hyNWcvSkNuWmJqTDlmRitGZWRrTk1lbWIKVXNsRXlEZVg4WWxKdTg0Y0tXVE1CM3BFTUdSY0N1L2Z3U0JheWh2ZElRS0JnUUQzcDd6ejBQU0U2M0h1a2xLRAo1L1RQRFVYaWlTaDYxMVBtRjE3ZHlGaUlEeDlLR2NjeDZHWEpFdkQxUVVCbU1tZ2tCVzR3Tmduc0dac0F1ZENoCk0vZituSUJZOHFUSGl6MDcydjMvbGg2eVFBdUkxRzZkd1c0TEIxbkVkWXNHQ2FEWFdwMnVjdENXZURKSEUyazAKamxhcGdYWEt6amJpUG95MUI5bGswWGFsRHdLQmdRRDlsUTY0NVEvdlZOY2hWT01vaXJQc0FFVGZWWWFRbEc0RwpYM2lsdXQ5V0xTZ0VGbUlXa2JFTDdBbS9wa0xlNkR0Z0JOVFkvaW9UY0Z3ZlpkQ3gzMUk2c3RXMjFiZkM4aE5UCkpJVEdaWnN3MzRNdkE0c2tIVldwU0Q1OFRwbzdmOUdxNDBRZGFId25NSXpZS0d3bVFhNExpUWRiTmVERlZaVWcKT0poV0lLQzluUUtCZ0hMT09xVFljQklTUjQydUhUeHkwaVVONVNEQUxwakxxMlFWMzI5MnlidkJMdU1lRWg2TAp1bXU5TVNDd3NCYndOU0RGdDYvM3pSKzdPd1RhSHVJc0FrSHc4UHJLK1NKYjVQWVRiZjRHN0pIN2FKMTB4UmF5CnJiM0ZHWk1lWG1WZEgwZldCUlR5dXNtV1ROQndLaXduOC9xK1BnblhFTkJPNk5HZkl1TSt5djUzQW9HQVdZQXgKcVFxWC9EWHFsRVljd3VOcHE0TlF5NjZwQndDTWNNMWpqTU1QTHFCVGZsWFVHdmtqRHYwcFY2NDc3ZWJ6aGVsTwpvcVBhZisvcFk5K1FzQ2NOSTlDWTQ2WTg3SjU2T1o4ckNjYzA1aVFOYmh1SEVhRVU2ZUdveEVFTjZMQXRkVzR2ClRWM2liQWFUdjN0TTZKZkg4U1JWSjFrMTB4Qm95TzJRdnQxWFBuVUNnWUJZUmFreUFMNlNxWVNwbWRldHg0TUEKRElRMDAyTVZyUFhrVyt2QmZwak1SUDRpNmRIdmNnR0tMWTBzM21uV3pqYkV5UmJPSENiUUpNTHdUQlVaS1kwNQpyeWtDMS9mc3lURm1wSU5yRHNQVWtkeVAwTWVWa1MvYTNuOWNxYTFOUWpBZ2E0aGFRdEt5N2ZtamVPeHdjcGIyCmJKbU02WXF6a1R3TG51bHR1MlFIQ3c9PQotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCkAwLjAuMC4wOjEyMzQ1"
)

func (kcp *KcpBase) HiddenConnListener() {
	serverPort := kcp.config.ServerPort - 1
	ColorL("hidden conn:", fmt.Sprintf("0.0.0.0:%d", serverPort))
	if listener, err := UseDefaultTlsConfig(fmt.Sprintf("0.0.0.0:%d", serverPort)).WithTlsListener(); err != nil {
		log.Println("control port start error:", err)
	} else {
		for {
			if con, err := listener.Accept(); err != nil {
				log.Println("control conn error, stop control con!")
				break
			} else {
				kcp.HiidenConfig(con)
			}
		}
	}
}

func (kcp *KcpBase) HiidenConfig(con net.Conn) {
	man := remote.NewApiServer(con)
	defer con.Close()

	con.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
	fromIP := strings.SplitN(con.RemoteAddr().String(), ":", 2)[0]
	user, _ := man.Talk(true, "user")
	if len(datas.MemDB.Kv) == 0 {
		if user == "user" {
			con.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
			if pwd, _ := man.Talk(true, "pwd"); pwd != "pwd" {
				return
			}
		}
	} else {
		pwdS := datas.MemDB.Kv[user]
		con.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
		if pwd, _ := man.Talk(true, "pwd"); pwd != pwdS {
			return
		}
	}

	for {
		con.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
		oper, err := man.Talk(false, "info", "user db", "forward to", "cancel forward", "set auth user/pwd", "exit")
		if err != nil {
			return
		}
		switch oper {
		case "info":
			conf := kcp.config
			data, _ := json.Marshal(conf)
			ColorL("Control", string(data))
			con.Write(data)
		case "user db":
			con.SetReadDeadline(time.Now().Add(time.Duration(300) * time.Second))
			man.DbCli()
		case "cancel forward":
			ips := []string{}
			for i := range kcp.RedirectBooks {
				ips = append(ips, i)
			}
			i, err := man.Choose("del which one:", ips...)
			if err != nil {
				return
			}
			datas.Locker.Lock()
			delete(kcp.RedirectBooks, i)
			datas.Locker.Unlock()
			if kcp.config.OldSSPwd != "" {
				kcp.config.SSPassword = kcp.config.OldSSPwd
			}
		case "forward to":
			con.SetReadDeadline(time.Now().Add(time.Duration(300) * time.Second))
			kcp.SetRedirectIRC(man, fromIP)
		case "set auth user/pwd":
			user, err := man.Input("set user name:")
			pwd, err := man.Input("set pwd name:")
			if err == nil {
				datas.MemDB.Kv[user] = pwd
			}
			con.Write([]byte("set user and pwd ok!"))
		case "exit":
			return
		}
	}
}

func (kcp *KcpBase) SetRedirectIRC(man *remote.ApiServerMan, ip string) {
	authName, err := man.Talk(true, "username")
	if err != nil {
		log.Println("talk error:", err)
		return
	}
	authPasswd, err := man.Talk(true, "pwd")
	if err != nil {
		log.Println("talk error:", err)
		return
	}
	defer os.RemoveAll("Kcpconfig")
	ColorL(authName, authPasswd)
	if configFile, err := Credient(authName, authPasswd); err != nil {
		log.Println("redirect error:", err)
		return
	} else {
		ColorL("Config file:", configFile)
		if IsDir(configFile) {
			book := NewBook()
			book.Scan(configFile)

			ColorL("scan file:", configFile)
			cs := book.GetServers()

			ColorL("scan :", len(cs))
			if err != nil {
				log.Println("talk error:", err)
				return
			}
			choose, err := man.Choose("One to Forward", cs...)
			if err != nil {
				log.Println("talk error:", err)
				return
			}
			version, err := man.Choose("Phone version / Pc", "Phone", "Pc normal")
			if err != nil {
				log.Println("talk time error:", err)
				return
			}
			if version != "" {
				t := 86400
				if choose != "" {
					keys := strings.SplitN(choose, ":", 2)
					conf := book.Get(keys[0])
					route := new(Route)
					if version == "Phone" {
						conf.ServerPort += 2
						conf.SALT = "kcp-go"
						conf.EBUFLEN = 4096
						kcp.config.OldSSPwd = kcp.config.SSPassword
						kcp.config.SSPassword = conf.Password

					} else {
						conf.SALT = "demo salt"
						conf.EBUFLEN = 1024
					}
					route.SetMode("proxy")
					route.SetExireTime(t)
					route.SetConfig(conf)
					chooseIps := []string{}
					for ip := range datas.MemDB.Kd["ip who use"] {
						chooseIps = append(chooseIps, ip)
					}
					ip, err := man.Choose("choose which ip use forward:", chooseIps...)
					if err != nil {
						return
					}
					kcp.RedirectBooks[ip] = route

					ColorL("redirect ->", route.config)
				}
			}

		}
	}
}

func UseDefaultTlsConfig(addr string) (tlsConfig *TlsConfig) {
	config := ParseURI(PASWD)

	ts := strings.SplitN(addr, ":", 2)
	config.Server = ts[0]
	config.ServerPort, _ = strconv.Atoi(ts[1])

	tlsConfig, err := config.ToTlsConfig()
	// fmt.Println("use :", config, tlsConfig, err)
	// fmt.Println(tlsConfig)
	if err != nil {
		log.Fatal("Create tls config failed: ", err)
	}
	return
}
