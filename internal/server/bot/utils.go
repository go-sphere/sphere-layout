package bot

import "github.com/go-sphere/sphere/social/telegram"

func NewButton[T any](text, query string, data T) telegram.Button {
	return telegram.NewButton(text, query, data)
}

func NewButtonX[T any](text string, extra *telegram.MethodExtraData, data T) telegram.Button {
	return telegram.NewButton(text, extra.CallbackQuery, data)
}

func UnmarshalUpdateData[T any](update *telegram.Update) (*T, error) {
	if update != nil && update.CallbackQuery != nil {
		_, data, err := telegram.UnmarshalData[T](update.CallbackQuery.Data)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return nil, nil
	}
}

func UnmarshalUpdateDataWithDefault[T any](update *telegram.Update, defaultValue *T) (*T, error) {
	if update != nil && update.CallbackQuery != nil {
		_, data, err := telegram.UnmarshalData[T](update.CallbackQuery.Data)
		if err != nil {
			if defaultValue != nil {
				return defaultValue, nil
			} else {
				return nil, err
			}
		}
		return data, nil
	} else {
		return defaultValue, nil
	}
}
