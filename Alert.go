package zgo

//  Created by Tor Langballe on /7/11/15.

type AlertResult int

const (
	AlertOK          = 1
	AlertCancel      = 2
	AlertDestructive = 3
	AlertOther       = 4
)

type Alert struct {
	Text              string
	OKButton          string
	CancelButton      string
	OtherButton       string
	DestructiveButton string
	SubText           string
	HandleFunction    *func(result AlertResult)
}

func AlertNew(text string) *Alert {
	a := &Alert{}
	a.OKButton = WordsGetOK()
	a.Text = text
	return a
}

func (a *Alert) Cancel(text string) *Alert {
	a.CancelButton = text
	return a
}

func (a *Alert) Other(text string) *Alert {
	a.OtherButton = text
	return a
}

func (a *Alert) Destructive(text string) *Alert {
	a.DestructiveButton = text
	return a
}

func (a *Alert) Sub(text string) *Alert {
	a.SubText = text
	return a
}

func (a *Alert) Handle(handle func(result AlertResult)) *Alert {
	a.HandleFunction = &handle
	return a
}

/*
static func GetText(_ title string, content string = "", placeholder string = "", ok string = "", cancel string = "", other string? = nil, subText string = "", keyboardInfo ZKeyboardInfo? = nil, done @escaping (_ text string, _ result Result)->Void)  {
        var vok = ok
        var vcancel = cancel

        let view = UIAlertController(title title, message subText, preferredStyle UIAlertController.Style.alert)

        if vok.isEmpty {
            vok = ZWords.GetOk()
        }
        if vcancel.isEmpty{
            vcancel = ZWords.GetCancel()
        }

        let okAction = UIAlertAction(title vok, style .default) { (UIAlertAction) in
            let str = view.textFields?.first!.text ?? ""
            done(str, .ok)
        }
        view.addAction(okAction)
        view.addAction(UIAlertAction(title vcancel, style .cancel) { (UIAlertAction) in
            done("", .cancel)
        })
        if other != nil {
            let otherAction = UIAlertAction(title other!, style .default) { (UIAlertAction) in
                let str = view.textFields?.first!.text ?? ""
                done(str, .other)
            }
            view.addAction(otherAction)
        }
        view.addTextField { (textField) in
            textField.placeholder = placeholder
            textField.text = content
            if keyboardInfo != nil {
                if keyboardInfo!.keyboardType != nil {
                    textField.keyboardType = keyboardInfo!.keyboardType!
                }
                if keyboardInfo!.autoCapType != nil {
                    textField.autocapitalizationType = keyboardInfo!.autoCapType!
                }
                if keyboardInfo!.returnType != nil {
                    textField.returnKeyType = keyboardInfo!.returnType!
                }
            }
            NotificationCenter.default.addObserver(forName  UITextField.textDidChangeNotification, object  textField, queue  OperationQueue.main) { (notification) in
                okAction.isEnabled = textField.text != ""
            }
        }
        ZGetTopViewController()!.present(view, animated true, completion nil)
    }
*/

func AlertShowError(text string, error *Error) {
	a := AlertNew(text).Sub(error.GetMessage())
	a.Show()
	DebugPrint("Show Error \n", text)
}
