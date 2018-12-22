using System;

namespace JsonTreeView
{
    [Flags]
    enum KeyStates
    {
        LeftButton = 1,
        RightButton = 2,
        Shift = 4,
        Control = 8,
        MiddleButton = 16,
        Alt = 32
    }
}
