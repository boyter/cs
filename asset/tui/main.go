// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	var textView *tview.TextView
	var dropdown *tview.DropDown
	var inputField *tview.InputField
	var extInputField *tview.InputField

	textView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true).
		ScrollToBeginning()

	//list := tview.NewList().
	//	AddItem("List item 1", "Some explanatory text", 'a', nil).
	//	AddItem("List item 2", "Some explanatory text", 'b', nil).
	//	AddItem("List item 3", "Some explanatory text", 'c', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", `Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text Some explanatory text `, 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("List item 4", "Some explanatory text", 'd', nil).
	//	AddItem("Quit", "Press to exit", 'q', func() {
	//		app.Stop()
	//	})

	dropdown = tview.NewDropDown().
		SetOptions([]string{"50", "100", "200", "300", "400", "500", "600", "700", "800", "900"}, nil).
		SetCurrentOption(3).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.Color16).
		SetSelectedFunc(func(text string, index int) {
			app.SetFocus(inputField)
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(inputField)
			}
		})

	extInputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(10).
		SetChangedFunc(func(text string) {

		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(dropdown)
			}
		})

	inputField = tview.NewInputField().
		SetFieldBackgroundColor(tcell.Color16).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetFieldWidth(0).
		SetChangedFunc(func(text string) {

		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyTab:
				app.SetFocus(extInputField)
			}
		})

	queryFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(inputField, 0, 8, false).
		AddItem(extInputField, 10, 0, false).
		AddItem(dropdown, 4, 1, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(queryFlex, 1, 0, false).
		AddItem(textView, 0, 3, false)
		//AddItem(list, 0, 3, false)

	_, _ = fmt.Fprintf(textView, "%s", `How slowly the time passes here, encompassed as I am by frost and snow!
Yet a second step is taken towards my enterprise.  I have hired a
vessel and am occupied in collecting my sailors; those whom I have
already engaged appear to be men on whom I can depend and are certainly
possessed of dauntless courage.

But I have one want which I have never yet been able to satisfy, and the
absence of the object of which I now feel as a most severe evil, I have no
friend, Margaret: when I am glowing with the enthusiasm of success, there
will be none to participate my joy; if I am assailed by disappointment, no
one will endeavour to sustain me in dejection. I shall commit my thoughts
to paper, it is true; but that is a poor medium for the communication of
feeling. I desire the company of a man who could sympathise with me, whose
eyes would reply to mine. You may deem me romantic, my dear sister, but I
bitterly feel the want of a friend. I have no one near me, gentle yet
courageous, possessed of a cultivated as well as of a capacious mind, whose
tastes are like my own, to approve or amend my plans. How would such a
friend repair the faults of your poor brother! I am too ardent in execution
and too impatient of difficulties. But it is a still greater evil to me
that I am self-educated: for the first fourteen years of my life I ran wild
on a common and read nothing but our Uncle Thomas’ books of voyages.
At that age I became acquainted with the celebrated poets of our own
country; but it was only when it had ceased to be in my power to derive its
most important benefits from such a conviction that I perceived the
necessity of becoming acquainted with more languages than that of my native
country. Now I am twenty-eight and am in reality more illiterate than many
schoolboys of fifteen. It is true that I have thought more and that my
daydreams are more extended and magnificent, but they want (as the painters
call it) _keeping;_ and I greatly need a friend who would have sense
enough not to despise me as romantic, and affection enough for me to
endeavour to regulate my mind.

Well, these are useless complaints; I shall certainly find no friend on the
wide ocean, nor even here in Archangel, among merchants and seamen. Yet
some feelings, unallied to the dross of human nature, beat even in these
rugged bosoms. My lieutenant, for instance, is a man of wonderful courage
and enterprise; he is madly desirous of glory, or rather, to word my phrase
more characteristically, of advancement in his profession. He is an
Englishman, and in the midst of national and professional prejudices,
unsoftened by cultivation, retains some of the noblest endowments of
humanity. I first became acquainted with him on board a whale vessel;
finding that he was unemployed in this city, I easily engaged him to assist
in my enterprise.

The master is a person of an excellent disposition and is remarkable in the
ship for his gentleness and the mildness of his discipline. This
circumstance, added to his well-known integrity and dauntless courage, made
me very desirous to engage him. A youth passed in solitude, my best years
spent under your gentle and feminine fosterage, has so refined the
groundwork of my character that I cannot overcome an intense distaste to
the usual brutality exercised on board ship: I have never believed it to be
necessary, and when I heard of a mariner equally noted for his kindliness
of heart and the respect and obedience paid to him by his crew, I felt
myself peculiarly fortunate in being able to secure his services. I heard
of him first in rather a romantic manner, from a lady who owes to him the
happiness of her life. This, briefly, is his story. Some years ago he loved
a young Russian lady of moderate fortune, and having amassed a considerable
sum in prize-money, the father of the girl consented to the match. He saw
his mistress once before the destined ceremony; but she was bathed in
tears, and throwing herself at his feet, entreated him to spare her,
confessing at the same time that she loved another, but that he was poor,
and that her father would never consent to the union. My generous friend
reassured the suppliant, and on being informed of the name of her lover,
instantly abandoned his pursuit. He had already bought a farm with his
money, on which he had designed to pass the remainder of his life; but he
bestowed the whole on his rival, together with the remains of his
prize-money to purchase stock, and then himself solicited the young
woman’s father to consent to her marriage with her lover. But the old
man decidedly refused, thinking himself bound in honour to my friend, who,
when he found the father inexorable, quitted his country, nor returned
until he heard that his former mistress was married according to her
inclinations. “What a noble fellow!” you will exclaim. He is
so; but then he is wholly uneducated: he is as silent as a Turk, and a kind
of ignorant carelessness attends him, which, while it renders his conduct
the more astonishing, detracts from the interest and sympathy which
otherwise he would command.

Yet do not suppose, because I complain a little or because I can
conceive a consolation for my toils which I may never know, that I am
wavering in my resolutions.  Those are as fixed as fate, and my voyage
is only now delayed until the weather shall permit my embarkation.  The
winter has been dreadfully severe, but the spring promises well, and it
is considered as a remarkably early season, so that perhaps I may sail
sooner than I expected.  I shall do nothing rashly:  you know me
sufficiently to confide in my prudence and considerateness whenever the
safety of others is committed to my care.

I cannot describe to you my sensations on the near prospect of my
undertaking.  It is impossible to communicate to you a conception of
the trembling sensation, half pleasurable and half fearful, with which
I am preparing to depart.  I am going to unexplored regions, to “the
land of mist and snow,” but I shall kill no albatross; therefore do not
be alarmed for my safety or if I should come back to you as worn and
woeful as the “Ancient Mariner.”  You will smile at my allusion, but I
will disclose a secret.  I have often attributed my attachment to, my
passionate enthusiasm for, the dangerous mysteries of ocean to that
production of the most imaginative of modern poets.  There is something
at work in my soul which I do not understand.  I am practically
industrious—painstaking, a workman to execute with perseverance and
labour—but besides this there is a love for the marvellous, a belief
in the marvellous, intertwined in all my projects, which hurries me out
of the common pathways of men, even to the wild sea and unvisited
regions I am about to explore.

But to return to dearer considerations. Shall I meet you again, after
having traversed immense seas, and returned by the most southern cape of
Africa or America?  I dare not expect such success, yet I cannot bear to
look on the reverse of the picture.  Continue for the present to write to
me by every opportunity: I may receive your letters on some occasions when
I need them most to support my spirits.  I love you very tenderly. 
Remember me with affection, should you never hear from me again.

Your affectionate brother,
  Robert Walton`)

	if err := app.SetRoot(flex, true).SetFocus(textView).Run(); err != nil {
		panic(err)
	}
}
