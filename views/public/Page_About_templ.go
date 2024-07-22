// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.747
package public

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Page_About() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<article class=\"w-6/12 mx-auto\"><section><h1>About</h1><p>Juniper is a simple web framework and CMS built in Go.</p></section><section><h2>Philosophy</h2><p>Juniper is designed to be simple and easy to use.</p><p>It is designed to be easy to extend and easy to maintain.</p></section><section><h2>Features</h2><ul><li>Simple routing</li><li>Templating</li><li>Middleware</li><li>SQLite</li><li>Database per model</li><li>Authentication</li><li>Authorization</li><li>File uploads</li><li>Static file serving</li><li>Logging</li><li>Configuration</li><li>AxB Testing</li><li>SEO</li><li>Websockets</li></ul></section></article>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}
