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

// Panel is one panel's rendered data.
type Panel struct {
	Name  string
	Title string
	Value string
	Note  string
	Rows  []Row
}

// Row is one line in a panel that shows a table.
type Row struct {
	Label string
	Value string
}

component PanelBody(p Panel) {
	{ if p.Rows != nil {
		<ui.Table>
			<ui.TableBody>
				{ for _, r := range p.Rows {
					<ui.TableRow>
						<ui.TableCell>{r.Label}</ui.TableCell>
						<ui.TableCell class="text-right">{r.Value}</ui.TableCell>
					</ui.TableRow>
				} }
			</ui.TableBody>
		</ui.Table>
	} else {
		<div class="flex items-baseline gap-3">
			<span class="text-3xl font-semibold">{p.Value}</span>
			<ui.Badge variant="secondary">{p.Note}</ui.Badge>
		</div>
	} }
}
