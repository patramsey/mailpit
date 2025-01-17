import * as esbuild from 'esbuild'
import pluginVue from 'esbuild-plugin-vue-next'
import { sassPlugin } from 'esbuild-sass-plugin'

const doWatch = process.env.WATCH == 'true' ? true : false;
const doMinify = process.env.MINIFY == 'true' ? true : false;

const ctx = await esbuild.context(
    {
        entryPoints: [
            "server/ui-src/app.js",
            "server/ui-src/docs.js"
        ],
        bundle: true,
        minify: doMinify,
        sourcemap: false,
        outdir: "server/ui/dist/",
        plugins: [pluginVue(), sassPlugin()],
        loader: {
            ".svg": "file",
            ".woff": "file",
            ".woff2": "file",
        },
        logLevel: "info"
    }
)

if (doWatch) {
    await ctx.watch()
} else {
    await ctx.rebuild()
    ctx.dispose()
}
