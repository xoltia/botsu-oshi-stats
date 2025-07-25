// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.857
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "time"

type ChartData struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

type TopVTuberWithAppearances struct {
	TopVTuber
	Appearances int
}

type TopVTuberWithDuration struct {
	TopVTuber
	Duration time.Duration
}

type TimelinePageModel struct {
	Type                  string
	UserProfilePictureURL string
	TopVTubersAppearances []TopVTuberWithAppearances
	TopVTubersDuration    []TopVTuberWithDuration
	Timeline              ChartData
}

func TimelinePage(model TimelinePageModel) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>OshiStats</title><link rel=\"icon\" type=\"image/png\" href=\"/static/icon-64.png\"><link rel=\"stylesheet\" href=\"/static/tailwind.css\"><script src=\"/static/htmx.min.js\"></script><script src=\"/static/chartjs.min.js\"></script></head><body class=\"min-h-screen bg-gradient-to-br from-neutral-800 via-neutral-900 to-neutral-800\"><header class=\"bg-neutral-900 border-b border-neutral-700\"><div class=\"container mx-auto flex items-center justify-between px-6 py-4\"><a href=\"/\" class=\"flex items-center gap-4\"><img src=\"/static/icon-240.png\" alt=\"OshiStats Icon\" class=\"w-10 h-10 rounded\"><div class=\"select-none font-semibold\"><h2 class=\"text-neutral-200 mb-0 text-sm/4\">Botsu</h2><h1 class=\"text-white text-xl/6\">OshiStats</h1></div></a> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if model.UserProfilePictureURL != "" {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<img src=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var2 string
			templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(model.UserProfilePictureURL)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 51, Col: 49}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "\" alt=\"Profile\" class=\"w-10 h-10 rounded-full border border-neutral-600 shadow-sm\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "</div></header><main class=\"container mx-auto p-1 md:p-6 text-white\"><section class=\"px-2 py-4 md:flex justify-center space-x-4\"><div class=\"my-4\"><h2 class=\"text-2xl font-bold mb-4\">Top By Appearances</h2><ul class=\"gap-4 grid grid-cols-2\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, v := range model.TopVTubersAppearances {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "<li class=\"flex items-center h-full\"><img src=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(v.AvatarURL)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 62, Col: 39}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "\" alt=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(v.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 62, Col: 52}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "\" class=\"w-10 h-10 rounded-full ml-2 mr-4 object-cover\"><div><div class=\"text-neutral-100\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var5 string
			templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(v.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 64, Col: 57}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "</div><div class=\"text-neutral-300 italic text-sm\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var6 string
			templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(v.Appearances)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 65, Col: 79}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, " videos</div></div></li>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "</ul></div><div class=\"my-4\"><h2 class=\"text-2xl font-bold mb-4\">Top By Duration</h2><ul class=\"gap-4 grid grid-cols-2\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, v := range model.TopVTubersDuration {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, "<li class=\"flex items-center h-full\"><img src=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var7 string
			templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(v.AvatarURL)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 76, Col: 39}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "\" alt=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var8 string
			templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(v.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 76, Col: 52}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, "\" class=\"w-10 h-10 rounded-full ml-2 mr-4 object-cover\"><div><div class=\"text-neutral-100\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var9 string
			templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(v.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 78, Col: 57}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "</div><div class=\"text-neutral-300 italic text-sm\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var10 string
			templ_7745c5c3_Var10, templ_7745c5c3_Err = templ.JoinStringErrs(v.Duration.Truncate(time.Second).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 79, Col: 107}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var10))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, "</div></div></li>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 16, "</ul></div></section><section class=\"px-2 lg:items-center flex flex-col\"><h2 class=\"text-2xl font-bold mb-4\">Total Watch Time</h2><div class=\"w-full lg:w-300 h-96 flex flex-col items-center\"><canvas id=\"timeline-chart\"></canvas></div><script>\n            (function() {\n              const formatMinutes = (m) => m >= 60 ?\n                `${(m/60).toFixed(1)}h` :\n                `${m}m`;\n              const ctx = document.getElementById('timeline-chart');\n              new Chart(ctx, {\n                type: 'bar',\n                data: {\n                  labels: ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Var11, templ_7745c5c3_Err := templruntime.ScriptContentOutsideStringLiteral(model.Timeline.Labels)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 100, Col: 50}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var11)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 17, ",\n                  datasets: [{\n                    data: ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Var12, templ_7745c5c3_Err := templruntime.ScriptContentOutsideStringLiteral(model.Timeline.Values)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/components/timeline.templ`, Line: 102, Col: 50}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ_7745c5c3_Var12)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 18, ",\n                    borderWidth: 1,\n                    backgroundColor: '#dc2626',\n                  }]\n                },\n                options: {\n                  scales: {\n                    y: {\n                      grid: {\n                        color: 'oklch(26.8% 0.007 34.298)'\n                      },\n                      ticks: {\n                        callback: function(value, index, ticks) {\n                          return formatMinutes(value);\n                        }\n                      },\n                    },\n                    x: {\n                      grid: {\n                        color: 'oklch(26.8% 0.007 34.298)'\n                      }\n                    }\n                  },\n                  plugins: {\n                    legend: {\n                      display: false\n                    },\n                    tooltip: {\n                      callbacks: {\n                        label: function(context) {\n                          return formatMinutes(context.parsed.y);\n                        }\n                      }\n                    }\n                  }\n                }\n              });\n            })();\n          </script></section></main></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
