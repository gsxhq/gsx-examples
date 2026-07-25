package views

import "github.com/gsxhq/gsx-examples/streaming-partial/ui"

// PanelShell is a panel before its data arrives: a card whose body is a marker
// region holding a skeleton. A <template for="{name}"> replaces the region's
// contents when the data lands.
component PanelShell(name string, title string) {
	<ui.Card>
		<ui.CardHeader><ui.CardTitle>{title}</ui.CardTitle></ui.CardHeader>
		<ui.CardContent>
			<?start name={name}>
				<ui.Skeleton class="h-24 w-full"/>
			<?end>
		</ui.CardContent>
	</ui.Card>
}
